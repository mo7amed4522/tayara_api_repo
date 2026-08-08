package export

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ExportApi struct{}

// streamCSV writes the CSV bytes as an attachment download.
func streamCSV(c *gin.Context, filename string, data []byte) {
	if data == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"msg": "no data"})
		return
	}
	c.Writer.Header().Set("Content-Type", "text/csv; charset=utf-8")
	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+filename)
	c.Writer.WriteHeader(http.StatusOK)
	_, _ = c.Writer.Write(data)
}

func csvFilename(prefix string) string {
	return prefix + "_" + time.Now().UTC().Format("20060102_150405") + ".csv"
}

// ExportUsers
// @Tags Export
// @Summary Admin exports users to CSV
// @Produce text/csv
// @Param status query string false "status filter"
// @Router /api/v1/admin/export/users [get]
func (ExportApi) Users(c *gin.Context) {
	data, err := exportService.ExportServices.Users(c.Query("status"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	streamCSV(c, csvFilename("users"), data)
}

// AgentsAdmin
// @Tags Export
// @Summary Admin exports agents to CSV
// @Produce text/csv
// @Param status query string false "status filter"
// @Router /api/v1/admin/export/agents [get]
func (ExportApi) Agents(c *gin.Context) {
	data, err := exportService.ExportServices.Agents(c.Query("status"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	streamCSV(c, csvFilename("agents"), data)
}

// Bookings
// @Tags Export
// @Summary Admin exports bookings to CSV
// @Produce text/csv
// @Param status query string false "booking status filter"
// @Router /api/v1/admin/export/bookings [get]
func (ExportApi) Bookings(c *gin.Context) {
	data, err := exportService.ExportServices.Bookings(c.Query("status"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	streamCSV(c, csvFilename("bookings"), data)
}

// Payments
// @Tags Export
// @Summary Admin exports payment transactions to CSV
// @Produce text/csv
// @Param status query string false "status filter"
// @Router /api/v1/admin/export/payments [get]
func (ExportApi) Payments(c *gin.Context) {
	data, err := exportService.ExportServices.Payments(c.Query("status"))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	streamCSV(c, csvFilename("payments"), data)
}