package router

import (
	"core-service/controllers"
	"core-service/middleware"
	"net/http"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func New(e *echo.Echo) *echo.Echo {
	e.GET("/healthcheck", healthcheck)
	e.GET("/api/core/swagger/*", echoSwagger.WrapHandler)

	//DATA
	e.POST("/api/core/ocr-id-card", controllers.OcrIDCard, middleware.AuthMiddleware)
	e.POST("/api/core/liveness-license", controllers.LivenessLicense, middleware.AuthMiddleware)
	e.GET("/api/core/liveness-score/:livenessID", controllers.LivenessScore, middleware.AuthMiddleware)
	e.POST("/api/core/face-comparison", controllers.FaceComparison, middleware.AuthMiddleware)
	e.POST("/api/core/pre-calculator", controllers.Precalculator)
	e.POST("/api/core/dopa", controllers.DOPA, middleware.AuthMiddleware)
	e.GET("/api/core/user-state", controllers.UserState, middleware.AuthMiddleware)
	e.GET("/api/core/user-state-v2", controllers.UserState, middleware.AuthMiddlewareV2)

	e.POST("/api/core/pre-calculator-v1", controllers.Precalculatorv1)

	//USER
	e.POST("/api/core/loancal", controllers.Loancal, middleware.AuthMiddleware)
	e.POST("/api/core/update-user-profile", controllers.UpdateUserProfile, middleware.AuthMiddleware)
	e.DELETE("/api/core/delete-user", controllers.DeleteUser, middleware.AuthMiddleware)
	e.POST("/api/core/file-upload-avatar", controllers.FileUploadAvatar, middleware.AuthMiddleware)
	e.POST("/api/core/insert-trans-liveness", controllers.InsertTransLiveness, middleware.AuthMiddleware)
	e.POST("/api/core/insert-trans-ocr", controllers.InsertTransOcr, middleware.AuthMiddleware)

	//COUNTRY
	e.POST("/api/core/province", controllers.GetProvince)
	e.POST("/api/core/district", controllers.GetDistrict)
	e.POST("/api/core/subdistrict", controllers.GetSubDistrict)

	//DOCUMENT
	e.POST("/api/core/document", controllers.Gendocument)
	e.POST("/api/core/documenttest", controllers.GendocumentTEST)
	e.POST("/api/core/emailsmtp", controllers.EmailSMTPTEST)

	return e
}

func healthcheck(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}
