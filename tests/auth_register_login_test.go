package tests

import (
	ssov1 "github.com/DenisPopkov/protos/gen/go/sso"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sso/tests/suite"
	"testing"
	"time"
)

const (
	emptyAppID = 0
	appID      = 1
	appSecret  = "test-secret"

	passDefaultLen = 10
)

func TestRegisterLogin_Login_HappyPath(t *testing.T) {
	ctx, st := suite.New(t)

	phoneNumber := gofakeit.Phone()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		PhoneNumber: phoneNumber,
		Password:    pass,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, respReg.GetUserId())

	respLogin, err := st.AuthClient.Login(ctx, &ssov1.LoginRequest{
		PhoneNumber: phoneNumber,
		Password:    pass,
		AppId:       appID,
	})
	require.NoError(t, err)

	token := respLogin.GetToken()
	require.NotEmpty(t, token)

	loginTime := time.Now()

	tokenParsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(appSecret), nil
	})
	require.NoError(t, err)

	claims, ok := tokenParsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	assert.Equal(t, respReg.GetUserId(), int64(claims["uid"].(float64)))
	assert.Equal(t, phoneNumber, claims["phoneNumber"].(string))
	assert.Equal(t, appID, int(claims["app_id"].(float64)))

	const deltaSeconds = 1

	// check if exp of token is in correct range, ttl get from st.Cfg.TokenTTL
	assert.InDelta(t, loginTime.Add(st.Cfg.TokenTTL).Unix(), claims["exp"].(float64), deltaSeconds)
}

func TestRegisterLogin_DuplicatedRegistration(t *testing.T) {
	ctx, st := suite.New(t)

	phoneNumber := gofakeit.Phone()
	pass := randomFakePassword()

	respReg, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		PhoneNumber: phoneNumber,
		Password:    pass,
	})
	require.NoError(t, err)
	require.NotEmpty(t, respReg.GetUserId())

	respReg, err = st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
		PhoneNumber: phoneNumber,
		Password:    pass,
	})
	require.Error(t, err)
	assert.Empty(t, respReg.GetUserId())
	assert.ErrorContains(t, err, "user already exists")
}

func TestRegister_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		phoneNumber string
		password    string
		expectedErr string
	}{
		{
			name:        "Register with Empty Password",
			phoneNumber: gofakeit.Phone(),
			password:    "",
			expectedErr: "password is required",
		},
		{
			name:        "Register with Empty PhoneNumber",
			phoneNumber: "",
			password:    randomFakePassword(),
			expectedErr: "phoneNumber is required",
		},
		{
			name:        "Register with Both Empty",
			phoneNumber: "",
			password:    "",
			expectedErr: "phoneNumber is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				PhoneNumber: tt.phoneNumber,
				Password:    tt.password,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)

		})
	}
}

func TestLogin_FailCases(t *testing.T) {
	ctx, st := suite.New(t)

	tests := []struct {
		name        string
		phoneNumber string
		password    string
		appID       int32
		expectedErr string
	}{
		{
			name:        "Login with Empty Password",
			phoneNumber: gofakeit.Phone(),
			password:    "",
			appID:       appID,
			expectedErr: "password is required",
		},
		{
			name:        "Login with Empty PhoneNumber",
			phoneNumber: "",
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "phoneNumber is required",
		},
		{
			name:        "Login with Both Empty PhoneNumber and Password",
			phoneNumber: "",
			password:    "",
			appID:       appID,
			expectedErr: "phoneNumber is required",
		},
		{
			name:        "Login with Non-Matching Password",
			phoneNumber: gofakeit.Phone(),
			password:    randomFakePassword(),
			appID:       appID,
			expectedErr: "invalid phoneNumber or password",
		},
		{
			name:        "Login without AppID",
			phoneNumber: gofakeit.Phone(),
			password:    randomFakePassword(),
			appID:       emptyAppID,
			expectedErr: "app_id is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := st.AuthClient.Register(ctx, &ssov1.RegisterRequest{
				PhoneNumber: gofakeit.Phone(),
				Password:    randomFakePassword(),
			})
			require.NoError(t, err)

			_, err = st.AuthClient.Login(ctx, &ssov1.LoginRequest{
				PhoneNumber: tt.phoneNumber,
				Password:    tt.password,
				AppId:       tt.appID,
			})
			require.Error(t, err)
			require.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func randomFakePassword() string {
	return gofakeit.Password(true, true, true, true, false, passDefaultLen)
}
