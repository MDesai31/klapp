package main

import (
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"

	"klapp/internal/models"
)

func (app *application) adminInvoices(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	invoices, total, err := app.invoices.List(page)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	const pageSize = 25
	totalPages := (total + pageSize - 1) / pageSize

	app.render(w, r, http.StatusOK, "admin_invoices.tmpl", templateData{
		Invoices:    invoices,
		CurrentPage: page,
		TotalPages:  totalPages,
		PrevPage:    page - 1,
		NextPage:    page + 1,
	})
}

func (app *application) adminInvoiceView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	inv, err := app.invoices.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "admin_invoice_view.tmpl", templateData{
		Invoice: &inv,
		Flash:   app.sessionManager.PopString(r.Context(), "flash"),
	})
}

func (app *application) adminInvoiceSubmit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	inv, err := app.invoices.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	if !inv.Reviewed {
		// Reviewed doesn't depend on the email going out - the admin has
		// seen the invoice either way - but a send failure is surfaced in
		// the flash so it isn't lost in the server log.
		flash := "Invoice emailed and marked reviewed."
		if emailErr := sendInvoiceEmail(&inv); emailErr != nil {
			app.logger.Error("sending invoice email", "error", emailErr, "invoice_id", id)
			flash = "Invoice marked reviewed, but the email FAILED to send - check the mail setup."
		}
		if err := app.invoices.SetReviewed(id); err != nil {
			app.serverError(w, r, err)
			return
		}
		app.sessionManager.Put(r.Context(), "flash", flash)
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/invoices/%d", id), http.StatusSeeOther)
}

func sendInvoiceEmail(inv *models.Invoice) error {
	subject := fmt.Sprintf("Invoice #%d — %s, House %s (%s)", inv.ID, inv.CustomerName, inv.HouseNumber, inv.Date)
	var body strings.Builder
	fmt.Fprintf(&body, "Invoice #%d\n", inv.ID)
	fmt.Fprintf(&body, "Date: %s\n", inv.Date)
	fmt.Fprintf(&body, "House: %s\n", inv.HouseNumber)
	fmt.Fprintf(&body, "Customer: %s\n", inv.CustomerName)
	fmt.Fprintf(&body, "Workers: %d\n", inv.NoOfWorkers)
	fmt.Fprintf(&body, "Arrived: %s  Left: %s\n", inv.TimeArrived, inv.TimeLeft)
	if len(inv.Jobs) > 0 {
		fmt.Fprintf(&body, "\nJobs:\n")
		for _, j := range inv.Jobs {
			fmt.Fprintf(&body, "  - %s\n", j)
		}
	}
	if len(inv.Materials) > 0 {
		fmt.Fprintf(&body, "\nMaterials:\n")
		for _, m := range inv.Materials {
			fmt.Fprintf(&body, "  - %s\n", m)
		}
	}
	if inv.Comments != "" {
		fmt.Fprintf(&body, "\nComments: %s\n", inv.Comments)
	}
	fmt.Fprintf(&body, "\nSubmitted by: %s\n", inv.WorkerName)

	msg := "To: mylawncut@aol.com\nSubject: " + subject + "\n\n" + body.String()

	cmd := exec.Command("msmtp", "--config=/etc/msmtprc", "--account=default", "mylawncut@aol.com")
	cmd.Stdin = strings.NewReader(msg)
	return cmd.Run()
}

func (app *application) adminJobDescriptions(w http.ResponseWriter, r *http.Request) {
	jobs, err := app.catalog.ListJobs()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	mats, err := app.catalog.ListMaterials()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.render(w, r, http.StatusOK, "admin_job_descriptions.tmpl", templateData{
		JobDescriptions: jobs,
		Materials:       mats,
		Flash:           app.sessionManager.PopString(r.Context(), "flash"),
	})
}

func (app *application) adminCreateJobDescription(w http.ResponseWriter, r *http.Request) {
	desc := strings.TrimSpace(r.PostFormValue("description"))
	if desc == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	if _, err := app.catalog.CreateJob(desc); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Job description added.")
	http.Redirect(w, r, "/admin/catalog", http.StatusSeeOther)
}

func (app *application) adminDeleteJobDescription(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}
	if err := app.catalog.DeleteJob(id); err != nil {
		app.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin/catalog", http.StatusSeeOther)
}

func (app *application) adminCreateMaterial(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PostFormValue("name"))
	unit := strings.TrimSpace(r.PostFormValue("unit"))
	priceStr := strings.TrimSpace(r.PostFormValue("price"))
	if name == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	price, _ := strconv.ParseFloat(priceStr, 64)
	if _, err := app.catalog.CreateMaterial(name, unit, price); err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Material added.")
	http.Redirect(w, r, "/admin/catalog", http.StatusSeeOther)
}

func (app *application) adminDeleteMaterial(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}
	if err := app.catalog.DeleteMaterial(id); err != nil {
		app.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/admin/catalog", http.StatusSeeOther)
}
