package models

import "database/sql"

type JobDescription struct {
	ID          int
	Description string
}

type Material struct {
	ID    int
	Name  string
	Unit  string
	Price float64
}

type CatalogModel struct {
	DB *sql.DB
}

func (m *CatalogModel) ListJobs() ([]JobDescription, error) {
	rows, err := m.DB.Query(`SELECT id, description FROM job_descriptions ORDER BY description`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []JobDescription
	for rows.Next() {
		var j JobDescription
		if err := rows.Scan(&j.ID, &j.Description); err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

// SearchJobs returns job description strings matching the query prefix (case-insensitive).
func (m *CatalogModel) SearchJobs(q string) ([]string, error) {
	rows, err := m.DB.Query(
		`SELECT description FROM job_descriptions WHERE description LIKE ? ORDER BY description LIMIT 10`,
		"%"+q+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (m *CatalogModel) CreateJob(description string) (int, error) {
	result, err := m.DB.Exec(`INSERT OR IGNORE INTO job_descriptions (description) VALUES (?)`, description)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (m *CatalogModel) DeleteJob(id int) error {
	_, err := m.DB.Exec(`DELETE FROM job_descriptions WHERE id = ?`, id)
	return err
}

func (m *CatalogModel) ListMaterials() ([]Material, error) {
	rows, err := m.DB.Query(`SELECT id, name, unit, price FROM materials ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Material
	for rows.Next() {
		var mat Material
		if err := rows.Scan(&mat.ID, &mat.Name, &mat.Unit, &mat.Price); err != nil {
			return nil, err
		}
		out = append(out, mat)
	}
	return out, rows.Err()
}

// SearchMaterials returns material name strings matching the query.
func (m *CatalogModel) SearchMaterials(q string) ([]string, error) {
	rows, err := m.DB.Query(
		`SELECT name FROM materials WHERE name LIKE ? ORDER BY name LIMIT 10`,
		"%"+q+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (m *CatalogModel) CreateMaterial(name, unit string, price float64) (int, error) {
	result, err := m.DB.Exec(
		`INSERT OR IGNORE INTO materials (name, unit, price) VALUES (?, ?, ?)`,
		name, unit, price,
	)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (m *CatalogModel) DeleteMaterial(id int) error {
	_, err := m.DB.Exec(`DELETE FROM materials WHERE id = ?`, id)
	return err
}

func (m *CatalogModel) UpdateMaterial(id int, name, unit string, price float64) error {
	result, err := m.DB.Exec(
		`UPDATE materials SET name = ?, unit = ?, price = ? WHERE id = ?`,
		name, unit, price, id,
	)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNoRecord
	}
	return nil
}
