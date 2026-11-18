package businesslogic

import (
	"context"
	"net/http"
	"os"
	"time"

	"relik-user-service/internal/auth"

	"github.com/ranggadablues/gosok/auth/jwt"
	"github.com/ranggadablues/gosok/auth/provider"
	"github.com/ranggadablues/gosok/common"
	"github.com/ranggadablues/gosok/constants"
	"github.com/ranggadablues/relik-proto-library/user-proto/pb"
)

type IFacebookBusinessLogic interface {
	FacebookLogin(ctx context.Context, req *pb.FacebookLoginRequest) (*pb.FacebookAuthResponse, error)
}

type FacebookBusinessLogic struct {
	pb.UnimplementedFacebookAuthServiceServer

	userBL func() IUserBusinessLogic
}

func NewFacebookBusinessLogic() IFacebookBusinessLogic {
	return &FacebookBusinessLogic{
		userBL: NewUserBusinessLogic,
	}
}

func (u *FacebookBusinessLogic) FacebookLogin(ctx context.Context, req *pb.FacebookLoginRequest) (*pb.FacebookAuthResponse, error) {
	code := req.GetCode()
	cfg := provider.FacebookConfig{
		GraphAPIURL: os.Getenv("FACEBOOK_GRAPH_API"),
		APIVersion:  os.Getenv("FACEBOOK_GRAPH_API_TOKEN_VERSION"),
		AppID:       os.Getenv("FACEBOOK_APP_ID"),
		AppSecret:   os.Getenv("FACEBOOK_APP_SECRET"),
		RedirectURI: os.Getenv("FACEBOOK_REDIRECT_URI"),
		HTTPClient:  &http.Client{},
	}
	fbProvider, err := provider.NewFacebookProvider(cfg)
	if err != nil {
		return nil, common.Error("facebook provider error", err)
	}

	fbProfile, err := fbProvider.Verify(ctx, code)
	if err != nil {
		return nil, common.Error("facebook error", err)
	}

	// 2️⃣ Find or create user
	user, userdetails, err := u.userBL().FindOrCreateUser(ctx, fbProfile.Email, fbProfile.Name, constants.FacebookAuth)
	if err != nil {
		return nil, common.Error("user error", err)
	}

	// 3️⃣ Issue JWT
	jwtToken, jwtRefreshToken, err := auth.Manager().GenerateTokenPair(ctx, jwt.UserClaims{
		UserID:   user.Id,
		Email:    user.Email,
		Name:     userdetails.Name,
		Provider: "facebook",
	})
	if err != nil {
		return nil, common.Error("token error", err)
	}

	return &pb.FacebookAuthResponse{
		Tokens: &pb.FacebookAuthTokens{
			AccessToken:      jwtToken,
			ExpiresAt:        time.Now().Add(15 * time.Minute).Unix(),
			RefreshToken:     jwtRefreshToken,
			RefreshExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
		},
		Profile: &pb.UserProfile{
			Id:    user.Id,
			Email: user.Email,
			Name:  userdetails.Name,
		},
	}, nil
}

