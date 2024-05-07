package server

import (
	"Homework/internal/domain"
	"Homework/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
)

const queryParamKey = "key"

type AddPickupPointRequest struct {
	Name           string `json:"name"`
	Address        string `json:"address"`
	ContactDetails string `json:"contact_details"`
}

type AddPickupPointResponse struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Address        string `json:"address"`
	ContactDetails string `json:"contact_details"`
}

type Server struct {
	Service *domain.BusinessService
}

func (s *Server) Create(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var request AddPickupPointRequest
	if err = json.Unmarshal(body, &request); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	pickupPointRepo := &repository.PickupPoint{
		Name:           request.Name,
		Address:        request.Address,
		ContactDetails: request.ContactDetails,
	}

	id, err := s.Service.Add(req.Context(), *pickupPointRepo)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := &AddPickupPointResponse{
		ID:             id,
		Name:           pickupPointRepo.Name,
		Address:        pickupPointRepo.Address,
		ContactDetails: pickupPointRepo.ContactDetails,
	}

	pickupPointJson, _ := json.Marshal(resp)
	w.Write(pickupPointJson)
}

func (s *Server) GetByID(w http.ResponseWriter, req *http.Request) {
	key, ok := mux.Vars(req)[queryParamKey]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	keyInt, err := strconv.ParseInt(key, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pickupPoint, err := s.Service.GetByID(req.Context(), keyInt)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	pickupPointJson, _ := json.Marshal(pickupPoint)
	w.Write(pickupPointJson)
}

func (s *Server) Update(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var request AddPickupPointRequest
	if err := json.Unmarshal(body, &request); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	idStr := mux.Vars(req)[queryParamKey]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pickupPoint := &repository.PickupPoint{
		ID:             id,
		Name:           request.Name,
		Address:        request.Address,
		ContactDetails: request.ContactDetails,
	}

	err = s.Service.Update(req.Context(), *pickupPoint)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) Delete(w http.ResponseWriter, req *http.Request) {
	idStr := mux.Vars(req)[queryParamKey]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.Service.Delete(req.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrObjectNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) List(w http.ResponseWriter, req *http.Request) {
	pickupPoints, err := s.Service.List(req.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	responseJSON, err := json.Marshal(pickupPoints)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(responseJSON)
}

func CreateRouter(implemetation Server) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/pickupPoint", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodPost:
			implemetation.Create(w, req)
		default:
			fmt.Println("url not found")
		}
	})

	router.HandleFunc(fmt.Sprintf("/pickupPoint/{%s:[0-9]+}", queryParamKey), func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			implemetation.GetByID(w, req)
		case http.MethodDelete:
			implemetation.Delete(w, req)
		case http.MethodPut:
			implemetation.Update(w, req)
		default:
			fmt.Println("url not found")
		}
	})

	router.HandleFunc("/pickupPoints", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			implemetation.List(w, req)
		default:
			fmt.Println("url not found")
		}
	})
	return router
}
