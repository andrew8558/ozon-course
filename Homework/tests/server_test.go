//go:build integration
// +build integration

package tests

import (
	"Homework/internal/repository"
	"Homework/internal/repository/postgresql"
	server1 "Homework/internal/server"
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_Add(t *testing.T) {
	pickupPoint := server1.AddPickupPointRequest{
		Name:           "pvz1",
		Address:        "spb",
		ContactDetails: "mail@mail.ru",
	}
	expectedPickupPoint := server1.AddPickupPointResponse{
		ID:             1,
		Name:           "pvz1",
		Address:        "spb",
		ContactDetails: "mail@mail.ru",
	}
	t.Run("add pickup point", func(t *testing.T) {
		//arrange
		db.SetUp(t, "pickup_points")
		repo := postgresql.NewPickupPoints(db.DB)
		implementation := server1.Server{Repo: repo}
		router := server1.CreateRouter(implementation)

		//act
		request, err := json.Marshal(pickupPoint)
		require.NoError(t, err)
		req := httptest.NewRequest("POST", "/pickupPoint", bytes.NewBuffer(request))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		//assert
		assert.Equal(t, http.StatusOK, w.Code)

		var response server1.AddPickupPointResponse
		json.NewDecoder(w.Body).Decode(&response)
		assert.Equal(t, expectedPickupPoint, response)
	})
}

func Test_Get(t *testing.T) {
	pickupPoint := server1.AddPickupPointRequest{
		Name:           "pvz1",
		Address:        "spb",
		ContactDetails: "mail@mail.ru",
	}
	expectedPickupPoint := repository.PickupPoint{
		ID:             1,
		Name:           "pvz1",
		Address:        "spb",
		ContactDetails: "mail@mail.ru",
	}
	t.Run("succes get pickup point", func(t *testing.T) {
		//arrange
		db.SetUp(t, "pickup_points")
		repo := postgresql.NewPickupPoints(db.DB)
		implementation := server1.Server{Repo: repo}
		router := server1.CreateRouter(implementation)

		request, err := json.Marshal(pickupPoint)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/pickupPoint", bytes.NewBuffer(request))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		//act
		getReq := httptest.NewRequest("GET", "/pickupPoint/1", nil)
		getW := httptest.NewRecorder()
		router.ServeHTTP(getW, getReq)

		//assert
		assert.Equal(t, http.StatusOK, getW.Code)

		var response repository.PickupPoint
		json.NewDecoder(getW.Body).Decode(&response)
		assert.Equal(t, expectedPickupPoint, response)
	})
	t.Run("fail to get pickup point", func(t *testing.T) {
		//arrange
		db.SetUp(t, "pickup_points")
		repo := postgresql.NewPickupPoints(db.DB)
		implementation := server1.Server{Repo: repo}
		router := server1.CreateRouter(implementation)

		request, err := json.Marshal(pickupPoint)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/pickupPoint", bytes.NewBuffer(request))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		//act
		getReq := httptest.NewRequest("GET", "/pickupPoint/2", nil)
		getW := httptest.NewRecorder()
		router.ServeHTTP(getW, getReq)

		//assert
		assert.Equal(t, http.StatusNotFound, getW.Code)
	})
}

func Test_Delete(t *testing.T) {
	pickupPoint := server1.AddPickupPointRequest{
		Name:           "pvz1",
		Address:        "spb",
		ContactDetails: "mail@mail.ru",
	}
	t.Run("succes delete pickup point", func(t *testing.T) {
		//arrange
		db.SetUp(t, "pickup_points")
		repo := postgresql.NewPickupPoints(db.DB)
		implementation := server1.Server{Repo: repo}
		router := server1.CreateRouter(implementation)

		request, err := json.Marshal(pickupPoint)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/pickupPoint", bytes.NewBuffer(request))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		//act
		delReq := httptest.NewRequest("DELETE", "/pickupPoint/1", nil)
		delW := httptest.NewRecorder()
		router.ServeHTTP(delW, delReq)

		//assert
		assert.Equal(t, http.StatusOK, delW.Code)
	})
	t.Run("fail to delete pickup point", func(t *testing.T) {
		//arrange
		db.SetUp(t, "pickup_points")
		repo := postgresql.NewPickupPoints(db.DB)
		implementation := server1.Server{Repo: repo}
		router := server1.CreateRouter(implementation)

		request, err := json.Marshal(pickupPoint)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/pickupPoint", bytes.NewBuffer(request))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		//act
		delReq := httptest.NewRequest("DELETE", "/pickupPoint/2", nil)
		delW := httptest.NewRecorder()
		router.ServeHTTP(delW, delReq)

		//assert
		assert.Equal(t, http.StatusNotFound, delW.Code)
	})
}

func Test_List(t *testing.T) {
	pickupPointsRequest := []server1.AddPickupPointRequest{server1.AddPickupPointRequest{
		Name:           "pvz1",
		Address:        "spb",
		ContactDetails: "mail@mail.ru",
	}, server1.AddPickupPointRequest{
		Name:           "pvz2",
		Address:        "msk",
		ContactDetails: "mail@mail.com",
	}}

	expectedPickupPoints := []repository.PickupPoint{repository.PickupPoint{
		ID:             1,
		Name:           "pvz1",
		Address:        "spb",
		ContactDetails: "mail@mail.ru",
	}, repository.PickupPoint{
		ID:             2,
		Name:           "pvz2",
		Address:        "msk",
		ContactDetails: "mail@mail.com",
	}}

	t.Run("get list pickup points", func(t *testing.T) {
		//arrange
		db.SetUp(t, "pickup_points")
		repo := postgresql.NewPickupPoints(db.DB)
		implementation := server1.Server{Repo: repo}
		router := server1.CreateRouter(implementation)

		for _, pickupPoint := range pickupPointsRequest {
			request, err := json.Marshal(pickupPoint)
			require.NoError(t, err)
			req := httptest.NewRequest("POST", "/pickupPoint", bytes.NewBuffer(request))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}

		//act
		getReq := httptest.NewRequest("GET", "/pickupPoints", nil)
		getW := httptest.NewRecorder()
		router.ServeHTTP(getW, getReq)

		//assert
		assert.Equal(t, http.StatusOK, getW.Code)

		var response []repository.PickupPoint
		json.NewDecoder(getW.Body).Decode(&response)
		assert.Equal(t, expectedPickupPoints, response)
	})
}

func Test_Update(t *testing.T) {
	pickupPoint := server1.AddPickupPointRequest{
		Name:           "pvz1",
		Address:        "spb",
		ContactDetails: "mail@mail.ru",
	}
	updatePickupPoint := server1.AddPickupPointRequest{
		Name:           "pvz1",
		Address:        "spb",
		ContactDetails: "yandex@yandex.ru",
	}
	t.Run("succes update pickup point", func(t *testing.T) {
		//arrange
		db.SetUp(t, "pickup_points")
		repo := postgresql.NewPickupPoints(db.DB)
		implementation := server1.Server{Repo: repo}
		router := server1.CreateRouter(implementation)

		request, err := json.Marshal(pickupPoint)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/pickupPoint", bytes.NewBuffer(request))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		//act
		request, err = json.Marshal(updatePickupPoint)
		require.NoError(t, err)
		updateReq := httptest.NewRequest("PUT", "/pickupPoint/1", bytes.NewBuffer(request))
		updateW := httptest.NewRecorder()
		router.ServeHTTP(updateW, updateReq)

		//assert
		assert.Equal(t, http.StatusOK, updateW.Code)
	})
	t.Run("fail to update pickup point", func(t *testing.T) {
		//arrange
		db.SetUp(t, "pickup_points")
		repo := postgresql.NewPickupPoints(db.DB)
		implementation := server1.Server{Repo: repo}
		router := server1.CreateRouter(implementation)

		request, err := json.Marshal(pickupPoint)
		require.NoError(t, err)

		req := httptest.NewRequest("POST", "/pickupPoint", bytes.NewBuffer(request))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		//act
		request, err = json.Marshal(pickupPoint)
		require.NoError(t, err)
		updateReq := httptest.NewRequest("PUT", "/pickupPoint/2", bytes.NewBuffer(request))
		updateW := httptest.NewRecorder()
		router.ServeHTTP(updateW, updateReq)

		//assert
		assert.Equal(t, http.StatusNotFound, updateW.Code)
	})
}
