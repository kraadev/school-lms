package meeting

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	Service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{Service: service}
}

func id(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.Param("id"))
}

func userID(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.GetString("user_id"))
}

func fail(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{"success": false, "message": err.Error()})
}

func (h *Handler) List(c *gin.Context) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	data, err := h.Service.List(uid, c.GetString("role"))
	if err != nil {
		fail(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func (h *Handler) Get(c *gin.Context) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	mid, err := id(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	data, err := h.Service.Get(uid, mid, c.GetString("role"), true)
	if err != nil {
		fail(c, http.StatusForbidden, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func (h *Handler) Create(c *gin.Context) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	var body CreateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	data, err := h.Service.Create(uid, body)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": data})
}

func (h *Handler) Update(c *gin.Context) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	mid, err := id(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	var body UpdateRequest
	if err := c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	data, err := h.Service.Update(uid, mid, body)
	if err != nil {
		fail(c, http.StatusForbidden, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func (h *Handler) Delete(c *gin.Context) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	mid, err := id(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	if err := h.Service.Delete(uid, mid, c.GetString("role")); err != nil {
		fail(c, http.StatusForbidden, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Meeting berhasil dihapus"})
}

func (h *Handler) Start(c *gin.Context) {
	h.status(c, true)
}

func (h *Handler) End(c *gin.Context) {
	h.status(c, false)
}

func (h *Handler) status(c *gin.Context, start bool) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	mid, err := id(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	var data interface{}
	if start {
		data, err = h.Service.Start(uid, mid)
	} else {
		data, err = h.Service.End(uid, mid)
	}
	if err != nil {
		fail(c, http.StatusForbidden, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

func (h *Handler) Request(c *gin.Context) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	mid, err := id(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	if err = h.Service.Request(uid, mid); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Pengajuan akses berhasil dikirim"})
}

func (h *Handler) Requests(c *gin.Context) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	mid, err := id(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	data, err := h.Service.Requests(uid, mid)
	if err != nil {
		fail(c, http.StatusForbidden, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

type decisionBody struct {
	Status string `json:"status" binding:"required"`
}

func (h *Handler) Decide(c *gin.Context) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	mid, err := id(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	target, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	var body decisionBody
	if err = c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	if err = h.Service.Decide(uid, mid, target, body.Status); err != nil {
		fail(c, http.StatusForbidden, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Keputusan peserta berhasil disimpan"})
}

func (h *Handler) DecideAll(c *gin.Context) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	mid, err := id(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	var body decisionBody
	if err = c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	if err = h.Service.DecideAll(uid, mid, body.Status); err != nil {
		fail(c, http.StatusForbidden, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Semua pengajuan peserta berhasil diproses"})
}

type permissionBody struct {
	MicEnabled         *bool `json:"mic_enabled"`
	CameraEnabled      *bool `json:"camera_enabled"`
	ScreenShareEnabled *bool `json:"screen_share_enabled"`
}

func (h *Handler) Permission(c *gin.Context) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	mid, err := id(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	target, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	var body permissionBody
	if err = c.ShouldBindJSON(&body); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	if err = h.Service.SetPermission(uid, mid, target, body.MicEnabled, body.CameraEnabled, body.ScreenShareEnabled); err != nil {
		fail(c, http.StatusForbidden, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Izin peserta berhasil diperbarui"})
}

func (h *Handler) Kick(c *gin.Context) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	mid, err := id(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	target, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	if err = h.Service.KickParticipant(uid, mid, target); err != nil {
		fail(c, http.StatusForbidden, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Peserta berhasil dikeluarkan"})
}

type muteBody struct {
	TrackSID string `json:"track_sid"`
}

func (h *Handler) Mute(c *gin.Context) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	mid, err := id(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	target, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	var body muteBody
	_ = c.ShouldBindJSON(&body)
	if err = h.Service.MuteParticipant(uid, mid, target, body.TrackSID); err != nil {
		fail(c, http.StatusForbidden, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Peserta berhasil dimute"})
}

func (h *Handler) Join(c *gin.Context) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	mid, err := id(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	role := c.GetString("role")
	resp, err := h.Service.Join(uid, mid, role)
	if err != nil {
		fail(c, http.StatusForbidden, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": resp})
}

func (h *Handler) Leave(c *gin.Context) {
	uid, err := userID(c)
	if err != nil {
		fail(c, http.StatusUnauthorized, err)
		return
	}
	mid, err := id(c)
	if err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	if err = h.Service.Leave(uid, mid); err != nil {
		fail(c, http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Berhasil keluar dari meeting"})
}
