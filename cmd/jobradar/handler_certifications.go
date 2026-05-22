package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/samnodier/jobradar/internal/convert"
	"github.com/samnodier/jobradar/internal/database"
	"github.com/samnodier/jobradar/internal/httpx"
)

type certificationCreateRequest struct {
	CertificationName   string  `json:"certification_name"`
	IssuingOrganization string  `json:"issuing_organization"`
	IssueDate           *string `json:"issue_date"`
	ExpirationDate      *string `json:"expiration_date"`
	DoesNotExpire       *bool   `json:"does_not_expire"`
	CredentialID        *string `json:"credential_id"`
	CredentialURL       *string `json:"credential_url"`
	IsInProgress        *bool   `json:"is_in_progress"`
}

type certificationUpdateRequest struct {
	CertificationName   *string `json:"certification_name"`
	IssuingOrganization *string `json:"issuing_organization"`
	IssueDate           *string `json:"issue_date"`
	ExpirationDate      *string `json:"expiration_date"`
	DoesNotExpire       *bool   `json:"does_not_expire"`
	CredentialID        *string `json:"credential_id"`
	CredentialURL       *string `json:"credential_url"`
	IsInProgress        *bool   `json:"is_in_progress"`
}

func (cfg *apiConfig) handlerCreateCertification(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	var req certificationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var issueDateStr string
	if req.IssueDate != nil {
		issueDateStr = *req.IssueDate
	}
	issueDate, err := convert.ToDateOptional(issueDateStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid issuing date: "+err.Error())
		return
	}

	var expirationDateStr string
	if req.ExpirationDate != nil {
		expirationDateStr = *req.ExpirationDate
	}
	expirationDate, err := convert.ToDateOptional(expirationDateStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid expiration date: "+err.Error())
		return
	}

	certification, err := cfg.db.CreateUserCertification(r.Context(), database.CreateUserCertificationParams{
		UserID:              userID,
		CertificationName:   req.CertificationName,
		IssuingOrganization: req.IssuingOrganization,
		IssueDate:           issueDate,
		ExpirationDate:      expirationDate,
		DoesNotExpire:       req.DoesNotExpire,
		CredentialID:        req.CredentialID,
		CredentialUrl:       req.CredentialURL,
		IsInProgress:        req.IsInProgress,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "certification not found or unauthorized")
			return
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "a certification with the details exist")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to create certification")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, certification)
}

func (cfg *apiConfig) handlerGetCertifications(w http.ResponseWriter, r *http.Request) {
	// User helper function to get the UserID
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	certifications, err := cfg.db.GetCertificationsByUserID(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, http.StatusInternalServerError, "failed to retrieve certifications")
		return
	}

	if certifications == nil {
		certifications = []database.GetCertificationsByUserIDRow{}
	}

	httpx.RespondJSON(w, http.StatusOK, certifications)
}

func (cfg *apiConfig) handlerUpdateCertification(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	certificationIDStr := chi.URLParam(r, "id")
	certificationID, err := uuid.Parse(certificationIDStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid certification ID format")
		return
	}

	var req certificationUpdateRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var issueDateStr string
	if req.IssueDate != nil {
		issueDateStr = *req.IssueDate
	}
	issueDate, err := convert.ToDateOptional(issueDateStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid issuing date: "+err.Error())
		return
	}

	var expirationDateStr string
	if req.ExpirationDate != nil {
		expirationDateStr = *req.ExpirationDate
	}
	expirationDate, err := convert.ToDateOptional(expirationDateStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid expiration date: "+err.Error())
		return
	}

	if req.DoesNotExpire != nil && !*req.DoesNotExpire && (req.ExpirationDate == nil || *req.ExpirationDate == "") {
		httpx.RespondError(w, http.StatusBadRequest, "expiration_date is required and does_not_expire is false")
		return
	}

	certification, err := cfg.db.UpdateUserCertification(r.Context(), database.UpdateUserCertificationParams{
		ID:                  certificationID,
		UserID:              userID,
		CertificationName:   req.CertificationName,
		IssuingOrganization: req.IssuingOrganization,
		IssueDate:           issueDate,
		ExpirationDate:      expirationDate,
		DoesNotExpire:       req.DoesNotExpire,
		CredentialID:        req.CredentialID,
		CredentialUrl:       req.CredentialURL,
		IsInProgress:        req.IsInProgress,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "certification not found or unauthorized")
			return
		}

		// check for postgres-specific errors
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				httpx.RespondError(w, http.StatusConflict, "a certification with the details already exists")
				return
			case "23503":
				httpx.RespondError(w, http.StatusBadRequest, "user account not found")
				return
			}
		}

		httpx.RespondError(w, http.StatusInternalServerError, "failed to update certification")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, certification)
}

func (cfg *apiConfig) handlerDeleteCertification(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromRequest(w, r)
	if !ok {
		return
	}

	certificationIDStr := chi.URLParam(r, "id")
	certificationID, err := uuid.Parse(certificationIDStr)
	if err != nil {
		httpx.RespondError(w, http.StatusBadRequest, "invalid certification ID format")
		return
	}

	err = cfg.db.DeleteUserCertification(r.Context(), database.DeleteUserCertificationParams{
		ID:     certificationID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.RespondError(w, http.StatusNotFound, "certification not found")
			return
		}
		httpx.RespondError(w, http.StatusInternalServerError, "failed to delete certification")
		return
	}

	w.WriteHeader(http.StatusNoContent) // 204 — success, no body
}
