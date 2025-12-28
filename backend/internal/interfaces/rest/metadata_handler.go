package rest

import (
	"github.com/nexuscrm/backend/internal/application/services"
)

type MetadataHandler struct {
	svc *services.ServiceManager
}

func NewMetadataHandler(svc *services.ServiceManager) *MetadataHandler {
	return &MetadataHandler{svc: svc}
}
