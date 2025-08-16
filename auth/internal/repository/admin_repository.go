// auth/internal/repository/admin_repo.go
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Language struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Subject struct {
	ID        string         `json:"id"`
	Slug      string         `json:"slug"`
	Name      map[string]any `json:"name"`
	CreatedAt time.Time      `json:"createdAt"`
}

type Direction struct {
	ID        string         `json:"id"`
	Slug      string         `json:"slug"`
	Name      map[string]any `json:"name"`
	CreatedAt time.Time      `json:"createdAt"`
}

type Subdirection struct {
	ID          string         `json:"id"`
	DirectionID string         `json:"directionId"`
	Slug        string         `json:"slug"`
	Name        map[string]any `json:"name"`
	CreatedAt   time.Time      `json:"createdAt"`
}

type TutorSubjectView struct {
	SubjectID  string `json:"subjectId"`
	Slug       string `json:"slug"`
	Level      string `json:"level"`
	PriceMinor int64  `json:"priceMinor"`
	Currency   string `json:"currency"`
}

type TutorLanguageView struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Proficiency string `json:"proficiency"`
}

type TutorSubdirectionView struct {
	SubdirectionID string `json:"subdirectionId"`
	Slug           string `json:"slug"`
	Level          string `json:"level"`
	PriceMinor     int64  `json:"priceMinor"`
	Currency       string `json:"currency"`
}

type AdminRepository interface {
	// taxonomies
	CreateLanguage(ctx context.Context, code, name string) error
	ListLanguages(ctx context.Context) ([]Language, error)

	CreateSubject(ctx context.Context, slug string, name map[string]any) error
	ListSubjects(ctx context.Context) ([]Subject, error)

	CreateDirection(ctx context.Context, slug string, name map[string]any) error
	ListDirections(ctx context.Context) ([]Direction, error)

	CreateSubdirection(ctx context.Context, directionID, directionSlug, slug string, name map[string]any) error
	ListSubdirections(ctx context.Context, directionSlug string) ([]Subdirection, error)

	// tutor bindings
	UpsertTutorSubject(ctx context.Context, tutorID, subjectID, subjectSlug, level string, price int64, currency string) error
	ListTutorSubjects(ctx context.Context, tutorID string) ([]TutorSubjectView, error)

	UpsertTutorLanguage(ctx context.Context, tutorID, code, proficiency string) error
	ListTutorLanguages(ctx context.Context, tutorID string) ([]TutorLanguageView, error)

	UpsertTutorSubdirection(ctx context.Context, tutorID, subdirID, subdirSlug, level string, price int64, currency string) error
	ListTutorSubdirections(ctx context.Context, tutorID string) ([]TutorSubdirectionView, error)
}

type adminRepo struct{ db *pgxpool.Pool }

func NewAdminRepository(db *pgxpool.Pool) AdminRepository { return &adminRepo{db: db} }

// ---------------- taxonomies ----------------

func (r *adminRepo) CreateLanguage(ctx context.Context, code, name string) error {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return fmt.Errorf("empty code")
	}
	_, err := r.db.Exec(ctx, `
INSERT INTO languages(code, name) VALUES ($1,$2)
ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name`, code, name)
	return err
}

func (r *adminRepo) ListLanguages(ctx context.Context) ([]Language, error) {
	rows, err := r.db.Query(ctx, `SELECT code, name FROM languages ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Language
	for rows.Next() {
		var l Language
		if err := rows.Scan(&l.Code, &l.Name); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, nil
}

func (r *adminRepo) CreateSubject(ctx context.Context, slug string, name map[string]any) error {
	slug = strings.ToLower(strings.TrimSpace(slug))
	b, _ := json.Marshal(name)
	_, err := r.db.Exec(ctx, `
INSERT INTO subjects(slug, name) VALUES ($1, $2::jsonb)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name`, slug, string(b))
	return err
}

func (r *adminRepo) ListSubjects(ctx context.Context) ([]Subject, error) {
	rows, err := r.db.Query(ctx, `SELECT id, slug, name, created_at FROM subjects ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Subject
	for rows.Next() {
		var s Subject
		var jb []byte
		if err := rows.Scan(&s.ID, &s.Slug, &jb, &s.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(jb, &s.Name)
		out = append(out, s)
	}
	return out, nil
}

func (r *adminRepo) CreateDirection(ctx context.Context, slug string, name map[string]any) error {
	slug = strings.ToLower(strings.TrimSpace(slug))
	b, _ := json.Marshal(name)
	_, err := r.db.Exec(ctx, `
INSERT INTO directions(slug, name) VALUES ($1,$2::jsonb)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name`, slug, string(b))
	return err
}

func (r *adminRepo) ListDirections(ctx context.Context) ([]Direction, error) {
	rows, err := r.db.Query(ctx, `SELECT id, slug, name, created_at FROM directions ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Direction
	for rows.Next() {
		var d Direction
		var jb []byte
		if err := rows.Scan(&d.ID, &d.Slug, &jb, &d.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(jb, &d.Name)
		out = append(out, d)
	}
	return out, nil
}

func (r *adminRepo) CreateSubdirection(ctx context.Context, directionID, directionSlug, slug string, name map[string]any) error {
	slug = strings.ToLower(strings.TrimSpace(slug))
	var dirID string
	if directionID != "" {
		dirID = directionID
	} else {
		// resolve by slug
		if err := r.db.QueryRow(ctx, `SELECT id FROM directions WHERE slug=$1`, strings.ToLower(directionSlug)).Scan(&dirID); err != nil {
			return fmt.Errorf("direction not found")
		}
	}
	b, _ := json.Marshal(name)
	_, err := r.db.Exec(ctx, `
INSERT INTO subdirections(direction_id, slug, name) VALUES ($1,$2,$3::jsonb)
ON CONFLICT (slug) DO UPDATE SET direction_id = EXCLUDED.direction_id, name = EXCLUDED.name`, dirID, slug, string(b))
	return err
}

func (r *adminRepo) ListSubdirections(ctx context.Context, directionSlug string) ([]Subdirection, error) {
	var rows pgxpool.Rows
	var err error
	if strings.TrimSpace(directionSlug) == "" {
		rows, err = r.db.Query(ctx, `SELECT id, direction_id, slug, name, created_at FROM subdirections ORDER BY created_at DESC`)
	} else {
		rows, err = r.db.Query(ctx, `
SELECT s.id, s.direction_id, s.slug, s.name, s.created_at
FROM subdirections s
JOIN directions d ON d.id = s.direction_id
WHERE d.slug = $1
ORDER BY s.created_at DESC`, strings.ToLower(directionSlug))
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Subdirection
	for rows.Next() {
		var x Subdirection
		var jb []byte
		if err := rows.Scan(&x.ID, &x.DirectionID, &x.Slug, &jb, &x.CreatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(jb, &x.Name)
		out = append(out, x)
	}
	return out, nil
}

// --------------- tutor bindings ---------------

func (r *adminRepo) UpsertTutorSubject(ctx context.Context, tutorID, subjectID, subjectSlug, level string, price int64, currency string) error {
	if subjectID == "" {
		if err := r.db.QueryRow(ctx, `SELECT id FROM subjects WHERE slug=$1`, strings.ToLower(subjectSlug)).Scan(&subjectID); err != nil {
			return fmt.Errorf("subject not found")
		}
	}
	if currency == "" {
		currency = "KZT"
	}
	_, err := r.db.Exec(ctx, `
INSERT INTO tutor_subjects (tutor_id, subject_id, level, price_minor, currency)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (tutor_id, subject_id) DO UPDATE
SET level=EXCLUDED.level, price_minor=EXCLUDED.price_minor, currency=EXCLUDED.currency`,
		tutorID, subjectID, nullIfEmpty(level), price, strings.ToUpper(currency))
	return err
}

func (r *adminRepo) ListTutorSubjects(ctx context.Context, tutorID string) ([]TutorSubjectView, error) {
	rows, err := r.db.Query(ctx, `
SELECT ts.subject_id, s.slug, COALESCE(ts.level,''), COALESCE(ts.price_minor,0), COALESCE(ts.currency,'KZT')
FROM tutor_subjects ts
JOIN subjects s ON s.id = ts.subject_id
WHERE ts.tutor_id = $1
ORDER BY s.slug`, tutorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TutorSubjectView
	for rows.Next() {
		var v TutorSubjectView
		if err := rows.Scan(&v.SubjectID, &v.Slug, &v.Level, &v.PriceMinor, &v.Currency); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func (r *adminRepo) UpsertTutorLanguage(ctx context.Context, tutorID, code, proficiency string) error {
	code = strings.ToLower(strings.TrimSpace(code))
	if code == "" {
		return fmt.Errorf("empty language code")
	}
	// seed if missing
	if _, err := r.db.Exec(ctx, `INSERT INTO languages(code,name) VALUES ($1,$2) ON CONFLICT (code) DO NOTHING`, code, code); err != nil {
		return err
	}
	prof := normalizeProficiency(proficiency)
	_, err := r.db.Exec(ctx, `
INSERT INTO tutor_languages (tutor_id, lang_code, proficiency)
VALUES ($1,$2,$3)
ON CONFLICT (tutor_id, lang_code) DO UPDATE SET proficiency = EXCLUDED.proficiency`,
		tutorID, code, prof)
	return err
}

func (r *adminRepo) ListTutorLanguages(ctx context.Context, tutorID string) ([]TutorLanguageView, error) {
	rows, err := r.db.Query(ctx, `
SELECT tl.lang_code, COALESCE(l.name,''), tl.proficiency
FROM tutor_languages tl
LEFT JOIN languages l ON l.code = tl.lang_code
WHERE tl.tutor_id = $1
ORDER BY tl.lang_code`, tutorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TutorLanguageView
	for rows.Next() {
		var v TutorLanguageView
		if err := rows.Scan(&v.Code, &v.Name, &v.Proficiency); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

func (r *adminRepo) UpsertTutorSubdirection(ctx context.Context, tutorID, subdirID, subdirSlug, level string, price int64, currency string) error {
	if subdirID == "" {
		if err := r.db.QueryRow(ctx, `SELECT id FROM subdirections WHERE slug=$1`, strings.ToLower(subdirSlug)).Scan(&subdirID); err != nil {
			return fmt.Errorf("subdirection not found")
		}
	}
	if currency == "" {
		currency = "KZT"
	}
	_, err := r.db.Exec(ctx, `
INSERT INTO tutor_subdirections (tutor_id, subdirection_id, level, price_minor, currency)
VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (tutor_id, subdirection_id) DO UPDATE
SET level=EXCLUDED.level, price_minor=EXCLUDED.price_minor, currency=EXCLUDED.currency`,
		tutorID, subdirID, nullIfEmpty(level), price, strings.ToUpper(currency))
	return err
}

func (r *adminRepo) ListTutorSubdirections(ctx context.Context, tutorID string) ([]TutorSubdirectionView, error) {
	rows, err := r.db.Query(ctx, `
SELECT ts.subdirection_id, sd.slug, COALESCE(ts.level,''), COALESCE(ts.price_minor,0), COALESCE(ts.currency,'KZT')
FROM tutor_subdirections ts
JOIN subdirections sd ON sd.id = ts.subdirection_id
WHERE ts.tutor_id = $1
ORDER BY sd.slug`, tutorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TutorSubdirectionView
	for rows.Next() {
		var v TutorSubdirectionView
		if err := rows.Scan(&v.SubdirectionID, &v.Slug, &v.Level, &v.PriceMinor, &v.Currency); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, nil
}

// helpers
func nullIfEmpty(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
func normalizeProficiency(p string) string {
	up := strings.ToUpper(strings.TrimSpace(p))
	if up == "NATIVE" {
		return "native"
	}
	switch up {
	case "A1", "A2", "B1", "B2", "C1", "C2":
		return up
	default:
		return "A1"
	}
}
