package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"klapp/internal/models"
)

func (app *application) adminCustomers(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))

	var (
		customers []models.Customer
		err       error
	)
	if q != "" {
		customers, err = app.customers.Search(q)
	} else {
		customers, err = app.customers.List()
	}
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "admin_customers.tmpl", templateData{
		Customers:     customers,
		SearchQuery:   q,
		Flash:         app.sessionManager.PopString(r.Context(), "flash"),
	})
}

func (app *application) adminCreateCustomer(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PostFormValue("name"))
	phone := strings.TrimSpace(r.PostFormValue("phone"))
	houseNumber := strings.TrimSpace(r.PostFormValue("house_number"))
	address := strings.TrimSpace(r.PostFormValue("address"))

	if name == "" || houseNumber == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if _, err := app.customers.Create(name, phone, houseNumber, address); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Customer added.")
	http.Redirect(w, r, "/admin/customers", http.StatusSeeOther)
}

func (app *application) adminCustomerView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	customer, err := app.customers.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	invoices, err := app.invoices.ListByCustomer(id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "admin_customer_view.tmpl", templateData{
		Customer: &customer,
		Invoices: invoices,
	})
}

func (app *application) adminEditCustomerForm(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	customer, err := app.customers.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	app.render(w, r, http.StatusOK, "admin_edit_customer.tmpl", templateData{Customer: &customer})
}

func (app *application) adminEditCustomer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		app.clientError(w, http.StatusNotFound)
		return
	}

	name := strings.TrimSpace(r.PostFormValue("name"))
	phone := strings.TrimSpace(r.PostFormValue("phone"))
	houseNumber := strings.TrimSpace(r.PostFormValue("house_number"))
	address := strings.TrimSpace(r.PostFormValue("address"))

	if name == "" || houseNumber == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err := app.customers.Update(id, name, phone, houseNumber, address); err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.clientError(w, http.StatusNotFound)
			return
		}
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/admin/customers/"+strconv.Itoa(id), http.StatusSeeOther)
}
