package businesslogic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ranggadablues/gosok/auth/jwt"
	"github.com/ranggadablues/gosok/auth/provider"
	"github.com/ranggadablues/gosok/constants"
	"github.com/ranggadablues/relik-proto-library/user-proto/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"relik-user-service/internal/auth"
)

type IGoogleBusinessLogic interface {
	GoogleLogin(ctx context.Context, req *pb.GoogleLoginRequest) (*pb.GoogleAuthResponse, error)
}

type GoogleBusinessLogic struct {
	pb.UnimplementedGoogleAuthServiceServer

	userBL func() IUserBusinessLogic
}

type googleUserInfoResponse struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func NewGoogleBusinessLogic() IGoogleBusinessLogic {
	return &GoogleBusinessLogic{
		userBL: NewUserBusinessLogic,
	}
}

func verifyGoogleAccessToken(ctx context.Context, accessToken string) (*googleUserInfoResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned status %d", resp.StatusCode)
	}

	var info googleUserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, err
	}

	if info.Email == "" || !info.EmailVerified {
		return nil, fmt.Errorf("unverified or missing google email")
	}

	return &info, nil
}

func (u *GoogleBusinessLogic) GoogleLogin(ctx context.Context, req *pb.GoogleLoginRequest) (*pb.GoogleAuthResponse, error) {
	var email string
	var name string

	// Verify Google token server-side in both flows
	if req.IdToken != "" {
		clientid := os.Getenv("GOOGLE_CLIENT_ID")
		googleProvider := provider.NewGoogleProvider(clientid)
		claims, err := googleProvider.Verify(ctx, req.IdToken)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid google token")
		}
		email = claims.Email
		name = claims.Name
	} else if req.AccessToken != "" {
		verifiedInfo, err := verifyGoogleAccessToken(ctx, req.AccessToken)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid google access token")
		}
		email = verifiedInfo.Email
		name = verifiedInfo.Name
		if req.UserInfo != nil {
			req.UserInfo.Email = verifiedInfo.Email
			req.UserInfo.Name = verifiedInfo.Name
			if verifiedInfo.Picture != "" {
				req.UserInfo.Picture = verifiedInfo.Picture
			}
		}
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "either id_token or access_token is required")
	}

	// 2️⃣ Find or create user
	user, userdetails, err := u.userBL().FindOrCreateUser(ctx, email, name, constants.GoogleAuth)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "user error: %v", err)
	}

	// 3️⃣ Issue JWT
	token, refreshToken, err := auth.Manager().GenerateTokenPair(ctx, jwt.UserClaims{
		UserID:   user.Id,
		Email:    user.Email,
		Name:     userdetails.Name,
		Provider: "google",
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "token error")
	}

	return &pb.GoogleAuthResponse{
		Tokens: &pb.GoogleAuthTokens{
			AccessToken:      token,
			ExpiresAt:        time.Now().Add(15 * time.Minute).Unix(),
			RefreshToken:     refreshToken,
			RefreshExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
		},
	}, nil
}
