package helper

import (
	"context"
	"fmt"
	"ucode/ucode_go_auth_service/api/models"

	"github.com/Timothylock/go-signin-with-apple/apple"
)



func GetAppleUserInfo(code string, c *models.AppleConfig) (*models.AppleUserPayload, error) {

	// Generate the client secret used to authenticate with Apple's validation servers
	secretKey := fmt.Sprintf(`-----BEGIN PRIVATE KEY-----
	%s
-----END PRIVATE KEY-----`, c.SecretKey)
	secret, err := apple.GenerateClientSecret(secretKey, c.TeamId, c.ClientId, c.KeyId)
	if err != nil {
		
		return nil, err
	}

	// Generate a new validation client
	client := apple.New()

	vReq := apple.AppValidationTokenRequest{
		ClientID:     c.ClientId,
		ClientSecret: secret,
		Code:         code,
	}

	var resp apple.ValidationResponse

	// Do the verification
	err = client.VerifyAppToken(context.Background(), vReq, &resp)
	if err != nil {
		
		return nil, err
	}

	if resp.Error != "" {
		fmt.Printf("apple returned an error: %s - %s\n", resp.Error, resp.ErrorDescription)
		return nil, fmt.Errorf("apple returned an error: %s - %s\n", resp.Error, resp.ErrorDescription)
	}

	// Get the unique user ID
	// unique, err := apple.GetUniqueID(resp.IDToken)
	// if err != nil {
	/
	// 	return nil, err
	// }

	// Get the email
	claim, err := apple.GetClaims(resp.IDToken)
	if err != nil {
		
		return nil, err
	}

	email := (*claim)["email"].(string)
	// emailVerified := (*claim)["email_verified"]
	// isPrivateEmail := (*claim)["is_private_email"]
	// name := (*claim)["name"]

	
	return &models.AppleUserPayload{
		Email: email,
	}, nil
}
