package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/ovh/cds/tools/smtpmock/sdk"
	"github.com/ovh/cds/tools/smtpmock/server/jwt"
	"github.com/ovh/cds/tools/smtpmock/server/store"
	"github.com/pkg/errors"
)

type Config struct {
	Port      int
	PortSMTP  int
	WithAuth  bool
	JwtSecret string
}

var config Config

func Start(ctx context.Context, c Config) error {
	config = c

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	if err := jwt.Init([]byte(config.JwtSecret)); err != nil {
		return err
	}

	e.GET("/", httpRootHandler)
	e.POST("/signin", httpSigninHandler)

	mess := e.Group("/messages", middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		Skipper: func(c echo.Context) bool {
			return !config.WithAuth
		},
		KeyLookup:  "header:" + echo.HeaderAuthorization,
		AuthScheme: "Bearer",
		Validator: func(key string, c echo.Context) (bool, error) {
			if _, err := jwt.CheckSessionToken(key); err != nil {
				return false, nil
			}
			return true, nil
		},
	}))

	{ // sub routes for /messages
		mess.GET("", func(c echo.Context) error {
			fmt.Println(c.Request().Header.Get("Authorization"))
			return c.JSON(http.StatusOK, store.GetMessages())
		})
		mess.GET("/:recipent", func(c echo.Context) error {
			return c.JSON(http.StatusOK, store.GetRecipientMessages(c.Param("recipent")))
		})
		mess.GET("/:recipent/latest", func(c echo.Context) error {
			messages := store.GetRecipientMessages(c.Param("recipent"))
			if len(messages) == 0 {
				return c.JSON(http.StatusNotFound, "not found")
			}
			return c.JSON(http.StatusOK, messages[0])
		})
	}

	return e.Start(fmt.Sprintf(":%d", config.Port))
}

func httpRootHandler(c echo.Context) error {
	var s = fmt.Sprintf("SMTP server listenning on %d\n", config.PortSMTP)
	s += fmt.Sprintf("%d mails received to %d recipents\n", store.CountMessages(), store.CountRecipients())
	return c.String(http.StatusOK, s)
}

func httpSigninHandler(c echo.Context) error {
	if !config.WithAuth {
		return c.JSON(http.StatusOK, sdk.SigninResponse{})
	}

	var data sdk.SigninRequest
	if err := c.Bind(&data); err != nil {
		return errors.WithStack(err)
	}

	subjectID, err := jwt.CheckSigninToken(data.SigninToken)
	if err != nil {
		return errors.WithStack(err)
	}

	sessionID, sessionToken, err := jwt.NewSessionToken(subjectID)
	if err != nil {
		return errors.WithStack(err)
	}

	store.AddSession(sessionID)

	return c.JSON(http.StatusOK, sdk.SigninResponse{
		SessionToken: sessionToken,
	})
}