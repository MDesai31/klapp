package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"klapp/internal/models"
)

// pinForm shows the PIN entry page. Language unknown at this point so default to Spanish.
func (app *application) pinForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "pin.tmpl", templateData{Spanish: true})
}

// pinSubmit verifies the PIN, stores the worker ID in the session, then redirects to the form.
func (app *application) pinSubmit(w http.ResponseWriter, r *http.Request) {
	pin := r.PostFormValue("pin")
	worker, err := app.workers.Authenticate(pin)
	if err != nil {
		if errors.Is(err, models.ErrInvalidPIN) {
			app.render(w, r, http.StatusOK, "pin.tmpl", templateData{
				Spanish: true,
				Flash:   "PIN no reconocido / PIN not recognized.",
			})
			return
		}
		app.serverError(w, r, err)
		return
	}

	app.session.Put(r.Context(), "workerID", worker.ID)
	http.Redirect(w, r, "/form", http.StatusSeeOther)
}

// invoiceForm renders the invoice submission form.
func (app *application) invoiceForm(w http.ResponseWriter, r *http.Request) {
	worker, ok := app.workerFromSession(w, r)
	if !ok {
		return
	}

	today := time.Now().Format("2006-01-02")
	app.render(w, r, http.StatusOK, "form.tmpl", templateData{
		Worker:          worker,
		Spanish:         worker.Language != "english",
		FormDate:        today,
		FormNoOfWorkers: "1",
		FormJobs:        []string{"", "", ""},
		FormMaterials:   []string{"", "", ""},
	})
}

// invoiceSubmit handles form submission.
func (app *application) invoiceSubmit(w http.ResponseWriter, r *http.Request) {
	worker, ok := app.workerFromSession(w, r)
	if !ok {
		return
	}
	spanish := worker.Language != "english"

	if err := r.ParseForm(); err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	date := strings.TrimSpace(r.FormValue("date"))
	houseNumber := strings.TrimSpace(r.FormValue("house_number"))
	customerName := strings.TrimSpace(r.FormValue("customer_name"))
	customerIDStr := strings.TrimSpace(r.FormValue("customer_id"))
	noOfWorkersStr := strings.TrimSpace(r.FormValue("no_of_workers"))
	timeArrived := strings.TrimSpace(r.FormValue("time_arrived"))
	timeLeft := strings.TrimSpace(r.FormValue("time_left"))
	comments := strings.TrimSpace(r.FormValue("comments"))

	rawJobs := r.Form["job"]
	rawMaterials := r.Form["material"]

	// Collect all form values for re-rendering on error.
	td := templateData{
		Worker:           worker,
		Spanish:          spanish,
		FormDate:         date,
		FormHouseNumber:  houseNumber,
		FormCustomerName: customerName,
		FormCustomerID:   customerIDStr,
		FormNoOfWorkers:  noOfWorkersStr,
		FormTimeArrived:  timeArrived,
		FormTimeLeft:     timeLeft,
		FormJobs:         padSlice(rawJobs, 3),
		FormMaterials:    padSlice(rawMaterials, 3),
		FormComments:     comments,
	}

	if date == "" || houseNumber == "" || noOfWorkersStr == "" || timeArrived == "" || timeLeft == "" {
		td.Flash = pick(spanish, "Por favor completa todos los campos requeridos.", "Please fill in all required fields.")
		app.render(w, r, http.StatusOK, "form.tmpl", td)
		return
	}

	noOfWorkers, err := strconv.Atoi(noOfWorkersStr)
	if err != nil || noOfWorkers < 1 {
		td.Flash = pick(spanish, "Número de trabajadores no válido.", "Invalid number of workers.")
		app.render(w, r, http.StatusOK, "form.tmpl", td)
		return
	}

	var customerID *int
	if customerIDStr != "" {
		if id, err := strconv.Atoi(customerIDStr); err == nil && id > 0 {
			customerID = &id
		}
	}

	jobs := filterEmpty(rawJobs)
	materials := filterEmpty(rawMaterials)

	_, err = app.invoices.Create(
		worker.ID, date, houseNumber, customerName, customerID,
		timeArrived, timeLeft, noOfWorkers, comments,
		jobs, materials,
	)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.session.Remove(r.Context(), "workerID")
	http.Redirect(w, r, "/success", http.StatusSeeOther)
}

func (app *application) successPage(w http.ResponseWriter, r *http.Request) {
	// Best-effort: show in Spanish since that's the majority default.
	// Session may already be cleared at this point.
	app.render(w, r, http.StatusOK, "success.tmpl", templateData{Spanish: true})
}

// apiCustomers returns JSON matching customers for a house number.
func (app *application) apiCustomers(w http.ResponseWriter, r *http.Request) {
	houseNumber := strings.TrimSpace(r.URL.Query().Get("house_number"))
	if houseNumber == "" {
		app.writeJSON(w, http.StatusOK, []any{})
		return
	}

	customers, err := app.customers.GetByHouseNumber(houseNumber)
	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, []any{})
		return
	}

	type customerResult struct {
		ID      int    `json:"id"`
		Name    string `json:"name"`
		Address string `json:"address"`
	}
	out := make([]customerResult, len(customers))
	for i, c := range customers {
		out[i] = customerResult{ID: c.ID, Name: c.Name, Address: c.Address}
	}
	app.writeJSON(w, http.StatusOK, out)
}

// apiJobs returns JSON job description suggestions matching q.
func (app *application) apiJobs(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		app.writeJSON(w, http.StatusOK, []string{})
		return
	}
	results, err := app.catalog.SearchJobs(q)
	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, []string{})
		return
	}
	if results == nil {
		results = []string{}
	}
	app.writeJSON(w, http.StatusOK, results)
}

// apiMaterials returns JSON material suggestions matching q.
func (app *application) apiMaterials(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		app.writeJSON(w, http.StatusOK, []string{})
		return
	}
	results, err := app.catalog.SearchMaterials(q)
	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, []string{})
		return
	}
	if results == nil {
		results = []string{}
	}
	app.writeJSON(w, http.StatusOK, results)
}

// workerFromSession retrieves the authenticated worker from the session.
// If there's no valid session it redirects to the PIN page and returns false.
func (app *application) workerFromSession(w http.ResponseWriter, r *http.Request) (*models.Worker, bool) {
	workerID, ok := app.session.Get(r.Context(), "workerID").(int)
	if !ok || workerID == 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil, false
	}
	worker, err := app.workers.Get(workerID)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return nil, false
	}
	return &worker, true
}

func pick(spanish bool, es, en string) string {
	if spanish {
		return es
	}
	return en
}

func filterEmpty(ss []string) []string {
	var out []string
	for _, s := range ss {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// padSlice ensures the slice has at least minLen elements (padding with empty strings).
func padSlice(ss []string, minLen int) []string {
	out := make([]string, len(ss))
	copy(out, ss)
	for len(out) < minLen {
		out = append(out, "")
	}
	return out
}
