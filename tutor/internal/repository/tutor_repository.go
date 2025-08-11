package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"tutor/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TutorRepository interface {
	// About
	UpsertAbout(ctx context.Context, userID string, first, last, phone, gender, avatarURL string, langs []domain.TutorLanguage) error
	// Availability
	ReplaceWeeklyAvailability(ctx context.Context, userID, timezone string, days []domain.AvailabilityDay) error
	// Education
	UpsertEducation(ctx context.Context, userID string, bio string, education []domain.Education, certs []domain.Certification) error
	// Subjects & prices
	ReplaceSubjects(ctx context.Context, userID string, items []domain.TutorSubjectDTO, regular map[string]int64, trial map[string]int64) error
	// Video
	SetVideo(ctx context.Context, userID string, videoURL string) error
	// Complete onboarding
	MarkCompleted(ctx context.Context, userID string) error

	// Queries
	FindTutorCardList(ctx context.Context, filters map[string]string, page, limit int) ([]domain.TutorProfile, int, error)
	FindTutorDetails(ctx context.Context, tutorID string) (*domain.TutorProfile, error)
}

type tutorRepository struct {
	db *pgxpool.Pool
}

func NewTutorRepository(db *pgxpool.Pool) TutorRepository {
	return &tutorRepository{db: db}
}

// ------------------ WRITE ------------------

func (r *tutorRepository) UpsertAbout(ctx context.Context, userID, first, last, phone, gender, avatarURL string, langs []domain.TutorLanguage) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// users (first/last/phone)
	if _, err := tx.Exec(ctx, `
UPDATE public.users
SET first_name = $2, last_name = $3, phone_e164 = $4, updated_at = now()
WHERE id = $1`, userID, first, last, nullIfEmpty(phone)); err != nil {
		return fmt.Errorf("update users: %w", err)
	}

	// tutor_profiles (ensure row)
	if _, err := tx.Exec(ctx, `
INSERT INTO public.tutor_profiles (user_id, hourly_rate_minor, currency, verification, props, created_at, updated_at)
VALUES ($1, 0, 'KZT', 'pending', '{}'::jsonb, now(), now())
ON CONFLICT (user_id) DO UPDATE SET updated_at = now()`, userID); err != nil {
		return fmt.Errorf("upsert tutor_profiles: %w", err)
	}

	// props: gender, avatar_url
	if _, err := tx.Exec(ctx, `
UPDATE public.tutor_profiles
SET props = COALESCE(props,'{}'::jsonb) || jsonb_build_object('gender', $2, 'avatar_url', $3),
    updated_at = now()
WHERE user_id = $1`, userID, nullIfEmpty(gender), nullIfEmpty(avatarURL)); err != nil {
		return fmt.Errorf("update props (about): %w", err)
	}

	// languages: replace all
	if _, err := tx.Exec(ctx, `DELETE FROM public.tutor_languages WHERE tutor_id = $1`, userID); err != nil {
		return err
	}
	for _, l := range langs {
		// если нет записи в languages(code) — можно вставить справочник заранее/миграциями
		if _, err := tx.Exec(ctx, `
INSERT INTO public.tutor_languages (tutor_id, lang_code, proficiency)
VALUES ($1,$2,$3)
ON CONFLICT (tutor_id, lang_code) DO UPDATE SET proficiency = EXCLUDED.proficiency`,
			userID, strings.ToLower(l.Code), strings.ToUpper(l.Proficiency)); err != nil {
			return fmt.Errorf("upsert tutor_languages: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (r *tutorRepository) ReplaceWeeklyAvailability(ctx context.Context, userID, timezone string, days []domain.AvailabilityDay) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// сохранить TZ в props
	if _, err := tx.Exec(ctx, `
UPDATE public.tutor_profiles
SET props = COALESCE(props,'{}'::jsonb) || jsonb_build_object('timezone', $2),
    updated_at = now()
WHERE user_id = $1`, userID, timezone); err != nil {
		return err
	}

	// чистим слоты (только рекуррентные для weekly мастера)
	if _, err := tx.Exec(ctx, `DELETE FROM public.availability_slots WHERE tutor_id = $1`, userID); err != nil {
		return err
	}

	// конвертим "Понедельник"+"18:00" → RRULE WEEKLY; BYDAY=MO; BYHOUR=18; BYMINUTE=0
	for _, d := range days {
		byday := mapDayRuToICal(d.Day)
		for _, t := range d.Slots {
			hh, mm := splitHHMM(t)
			rrule := fmt.Sprintf("FREQ=WEEKLY;BYDAY=%s;BYHOUR=%d;BYMINUTE=%d", byday, hh, mm)
			// фиктивный интервал 30 минут (UTC вычислит воркер при материализации календаря)
			start := time.Now().UTC()
			end := start.Add(30 * time.Minute)
			if _, err := tx.Exec(ctx, `
INSERT INTO public.availability_slots (id, tutor_id, starts_at, ends_at, recurrence, is_recurring, created_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, true, now())`, userID, start, end, rrule); err != nil {
				return fmt.Errorf("insert slot: %w", err)
			}
		}
	}

	return tx.Commit(ctx)
}

func (r *tutorRepository) UpsertEducation(ctx context.Context, userID string, bio string, education []domain.Education, certs []domain.Certification) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	eduBytes, _ := json.Marshal(education)
	certBytes, _ := json.Marshal(certs)

	if _, err := tx.Exec(ctx, `
UPDATE public.tutor_profiles
SET bio = $2,
    props = COALESCE(props,'{}'::jsonb) || jsonb_build_object('education', $3::jsonb, 'certificates', $4::jsonb),
    updated_at = now()
WHERE user_id = $1`, userID, bio, string(eduBytes), string(certBytes)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *tutorRepository) ReplaceSubjects(ctx context.Context, userID string, items []domain.TutorSubjectDTO, regular map[string]int64, trial map[string]int64) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// чистим предметы
	if _, err := tx.Exec(ctx, `DELETE FROM public.tutor_subjects WHERE tutor_id = $1`, userID); err != nil {
		return err
	}
	// вставляем новые
	for _, it := range items {
		if _, err := tx.Exec(ctx, `
INSERT INTO public.tutor_subjects (tutor_id, subject_id, level, price_minor, currency)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (tutor_id, subject_id) DO UPDATE
SET level = EXCLUDED.level, price_minor = EXCLUDED.price_minor, currency = EXCLUDED.currency`,
			userID, it.SubjectID, nullIfEmpty(it.Level), it.PriceMinor, strings.ToUpper(it.Currency)); err != nil {
			return fmt.Errorf("upsert tutor_subjects: %w", err)
		}
	}

	// сохраняем цены в props (trial)
	props := map[string]interface{}{"prices": map[string]interface{}{}}
	for slug, p := range regular {
		ensureNestedMap(props, "prices")[slug] = map[string]interface{}{
			"regular_minor": p,
			"trial_minor":   trial[slug],
		}
	}
	propsJSON, _ := json.Marshal(props)

	if _, err := tx.Exec(ctx, `
UPDATE public.tutor_profiles
SET props = COALESCE(props,'{}'::jsonb) || $2::jsonb,
    updated_at = now()
WHERE user_id = $1`, userID, string(propsJSON)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *tutorRepository) SetVideo(ctx context.Context, userID string, videoURL string) error {
	_, err := r.db.Exec(ctx, `
UPDATE public.tutor_profiles
SET video_url = $2, updated_at = now()
WHERE user_id = $1`, userID, videoURL)
	return err
}

func (r *tutorRepository) MarkCompleted(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `
UPDATE public.tutor_profiles
SET props = COALESCE(props,'{}'::jsonb) || jsonb_build_object('onboarding_completed_at', to_char(now(),'YYYY-MM-DD"T"HH24:MI:SSZ')),
    updated_at = now()
WHERE user_id = $1`, userID)
	return err
}

// ------------------ READ ------------------

func (r *tutorRepository) FindTutorCardList(ctx context.Context, filters map[string]string, page, limit int) ([]domain.TutorProfile, int, error) {
	// упрощённый листинг карточек
	offset := (page - 1) * limit
	// фильтры search по био/фио
	search := strings.TrimSpace(filters["search"])

	var rowsCount int
	if err := r.db.QueryRow(ctx, `
SELECT COUNT(*) FROM public.tutor_profiles tp
JOIN public.users u ON u.id = tp.user_id
WHERE tp.deleted_at IS NULL AND u.deleted_at IS NULL
  AND tp.verification IN ('pending','verified')
  AND ($1 = '' OR (tp.bio ILIKE '%'||$1||'%' OR u.first_name ILIKE '%'||$1||'%' OR u.last_name ILIKE '%'||$1||'%'))`,
		search).Scan(&rowsCount); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(ctx, `
SELECT tp.user_id, COALESCE(u.first_name,''), COALESCE(u.last_name,''), COALESCE(u.phone_e164,''),
       COALESCE(tp.props->>'gender',''), COALESCE(tp.props->>'avatar_url',''),
       COALESCE(tp.bio,''), COALESCE(tp.video_url,''), COALESCE(tp.props->>'timezone',''),
       COALESCE(tp.rating_avg,0), COALESCE(tp.rating_count,0), tp.verification, tp.created_at, tp.updated_at
FROM public.tutor_profiles tp
JOIN public.users u ON u.id = tp.user_id
WHERE tp.deleted_at IS NULL AND u.deleted_at IS NULL
  AND tp.verification IN ('pending','verified')
  AND ($1 = '' OR (tp.bio ILIKE '%'||$1||'%' OR u.first_name ILIKE '%'||$1||'%' OR u.last_name ILIKE '%'||$1||'%'))
ORDER BY tp.rating_avg DESC NULLS LAST, tp.updated_at DESC
LIMIT $2 OFFSET $3`, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := make([]domain.TutorProfile, 0, limit)
	for rows.Next() {
		var p domain.TutorProfile
		if err := rows.Scan(
			&p.UserID, &p.FirstName, &p.LastName, &p.PhoneE164,
			&p.Gender, &p.AvatarURL, &p.Bio, &p.VideoURL, &p.Timezone,
			&p.RatingAvg, &p.RatingCount, &p.Verification, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, p)
	}
	return list, rowsCount, nil
}

func (r *tutorRepository) FindTutorDetails(ctx context.Context, tutorID string) (*domain.TutorProfile, error) {
	row := r.db.QueryRow(ctx, `
SELECT tp.user_id, COALESCE(u.first_name,''), COALESCE(u.last_name,''), COALESCE(u.phone_e164,''),
       COALESCE(tp.props->>'gender',''), COALESCE(tp.props->>'avatar_url',''),
       COALESCE(tp.bio,''), COALESCE(tp.video_url,''), COALESCE(tp.props->>'timezone',''),
       COALESCE(tp.props->'education','[]'::jsonb),
       COALESCE(tp.props->'certificates','[]'::jsonb),
       COALESCE(tp.props->'prices','{}'::jsonb),
       COALESCE(tp.rating_avg,0), COALESCE(tp.rating_count,0), tp.verification, tp.created_at, tp.updated_at
FROM public.tutor_profiles tp
JOIN public.users u ON u.id = tp.user_id
WHERE tp.user_id = $1 AND tp.deleted_at IS NULL AND u.deleted_at IS NULL
`, tutorID)

	var p domain.TutorProfile
	var eduJSON, certJSON, pricesJSON []byte
	if err := row.Scan(
		&p.UserID, &p.FirstName, &p.LastName, &p.PhoneE164,
		&p.Gender, &p.AvatarURL, &p.Bio, &p.VideoURL, &p.Timezone,
		&eduJSON, &certJSON, &pricesJSON,
		&p.RatingAvg, &p.RatingCount, &p.Verification, &p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		return nil, err
	}
	_ = json.Unmarshal(eduJSON, &p.Education)
	_ = json.Unmarshal(certJSON, &p.Certificates)
	prices := map[string]map[string]int64{}
	_ = json.Unmarshal(pricesJSON, &prices)
	p.Prices, p.TrialPrices = map[string]int64{}, map[string]int64{}
	for slug, v := range prices {
		if v == nil {
			continue
		}
		if reg, ok := v["regular_minor"]; ok {
			p.Prices[slug] = reg
		}
		if tr, ok := v["trial_minor"]; ok {
			p.TrialPrices[slug] = tr
		}
	}

	// languages
	lrows, err := r.db.Query(ctx, `
SELECT tl.lang_code, tl.proficiency, l.name
FROM public.tutor_languages tl
LEFT JOIN public.languages l ON l.code = tl.lang_code
WHERE tl.tutor_id = $1`, tutorID)
	if err == nil {
		defer lrows.Close()
		for lrows.Next() {
			var lang domain.TutorLanguage
			if err := lrows.Scan(&lang.Code, &lang.Proficiency, &lang.DisplayName); err == nil {
				p.Languages = append(p.Languages, lang)
			}
		}
	}

	// subjects
	srows, err := r.db.Query(ctx, `
SELECT ts.subject_id, s.slug, COALESCE(ts.level,''), COALESCE(ts.price_minor,0), COALESCE(ts.currency,'KZT')
FROM public.tutor_subjects ts
JOIN public.subjects s ON s.id = ts.subject_id
WHERE ts.tutor_id = $1`, tutorID)
	if err == nil {
		defer srows.Close()
		for srows.Next() {
			var it domain.TutorSubjectDTO
			if err := srows.Scan(&it.SubjectID, &it.SubjectSlug, &it.Level, &it.PriceMinor, &it.Currency); err == nil {
				p.Subjects = append(p.Subjects, it)
			}
		}
	}

	return &p, nil
}

// ------------------ helpers ------------------

func mapDayRuToICal(day string) string {
	switch strings.ToLower(strings.TrimSpace(day)) {
	case "понедельник":
		return "MO"
	case "вторник":
		return "TU"
	case "среда":
		return "WE"
	case "четверг":
		return "TH"
	case "пятница":
		return "FR"
	case "суббота":
		return "SA"
	case "воскресенье":
		return "SU"
	default:
		return "MO"
	}
}
func splitHHMM(s string) (int, int) {
	var hh, mm int
	fmt.Sscanf(s, "%d:%d", &hh, &mm)
	return hh, mm
}
func nullIfEmpty(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	v := strings.TrimSpace(s)
	return &v
}
func ensureNestedMap(m map[string]interface{}, key string) map[string]interface{} {
	if _, ok := m[key]; !ok {
		m[key] = map[string]interface{}{}
	}
	return m[key].(map[string]interface{})
}
