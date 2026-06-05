package controllers

import (
	"monitoring-service/app/usecases"
	"net/http"
	"github.com/labstack/echo/v4"
)

type PendudukRiskController struct {
	usecase usecases.PendudukRiskUsecase
}

func NewPendudukRiskController(uc usecases.PendudukRiskUsecase) *PendudukRiskController {
	return &PendudukRiskController{usecase: uc}
}

// GetPendudukByRisk handles GET /api/penduduk-by-risk
func (ctrl *PendudukRiskController) GetPendudukByRisk(c echo.Context) error {
	kategori := c.QueryParam("kategori")
	risiko := c.QueryParam("risiko") // optional: Tinggi, Sedang, Rendah

	if kategori == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Parameter 'kategori' wajib diisi (anak, remaja, dewasa, lansia)",
		})
	}

	// Validasi nilai risiko jika diberikan
	if risiko != "" {
		validRisk := map[string]bool{"Tinggi": true, "Sedang": true, "Normal": true}
		if !validRisk[risiko] {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Parameter 'risiko' harus Tinggi, Sedang, atau Normal",
			})
		}
	}

	result, err := ctrl.usecase.GetPendudukByRisk(kategori, risiko)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}