package security

import (
	"errors"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// GenerateJWT ...
func GenerateJWT(m map[string]interface{}, tokenExpireTime time.Duration, tokenSecretKey string) (tokenString string, err error) {
	var token *jwt.Token

	token = jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	for key, value := range m {
		claims[key] = value
	}

	claims["iat"] = time.Now().Unix()
	claims["exp"] = time.Now().Add(tokenExpireTime).Unix()

	tokenString, err = token.SignedString([]byte(tokenSecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ExtractClaims extracts claims from given token
func ExtractClaims(tokenString string, tokenSecretKey string) (jwt.MapClaims, error) {
	var (
		token *jwt.Token
		err   error
	)

	token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// check token signing method etc
		return []byte(tokenSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !(ok && token.Valid) {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ExtractToken checks and returns token part of input string
func ExtractToken(bearer string) (token string, err error) {
	strArr := strings.Split(bearer, " ")
	if len(strArr) == 2 {
		return strArr[1], nil
	}
	return token, errors.New("wrong token format")
}

type TokenInfo struct {
	ID     string
	Tables []Table
	// ProjectID        string
	// ClientPlatformID string
	// ClientTypeID     string
	// UserID           string
	RoleID           string
	// IP               string
	// Data             string
	LoginTableSlug string
}

type Table struct {
	TableSlug string
	ObjectID  string
}

func ParseClaims(token string, secretKey string) (result TokenInfo, err error) {
	var ok bool
	var claims jwt.MapClaims

	claims, err = ExtractClaims(token, secretKey)
	if err != nil {
		return result, err
	}
	result.ID, ok = claims["id"].(string)
	result.RoleID = claims["role_id"].(string)
	if !ok {
		err = errors.New("cannot parse 'id' field")
		return result, err
	}
	if claims["tables"] != nil {
		for _, item := range claims["tables"].([]interface{}) {
			var table Table
			table.ObjectID = item.(map[string]interface{})["object_id"].(string)
			table.TableSlug = item.(map[string]interface{})["table_slug"].(string)
			result.Tables = append(result.Tables, table)
		}
	}
	loginTableSlug, ok := claims["login_table_slug"]
	if ok {
		result.LoginTableSlug = loginTableSlug.(string)
	}

	// projectID := claims["project_id"].(string)
	// clientPlatformID := claims["client_platform_id"].(string)
	// clientTypeID := claims["client_type_id"].(string)
	// userID := claims["user_id"].(string)
	// roleID := claims["role_id"].(string)
	// ip := claims["ip"].(string)
	// data := claims["data"].(string)

	return
}

type PasscodeTokenInfo struct {
	Phone      string
	HashedCode string
	ExpiresAt  string
	CreatedAt  string
}

func ParsePasscodeClaims(token string, secretKey string) (result PasscodeTokenInfo, err error) {
	var ok bool
	var claims jwt.MapClaims

	claims, err = ExtractClaims(token, secretKey)
	if err != nil {
		return result, err
	}

	result.Phone, ok = claims["phone"].(string)
	if !ok {
		err = errors.New("cannot parse 'phone' field")
		return result, err
	}

	result.HashedCode, ok = claims["hashed_code"].(string)
	if !ok {
		err = errors.New("cannot parse 'hashed_code' field")
		return result, err
	}

	result.ExpiresAt, ok = claims["expires_at"].(string)
	if !ok {
		err = errors.New("cannot parse 'expires_at' field")
		return result, err
	}

	result.CreatedAt, ok = claims["created_at"].(string)
	if !ok {
		err = errors.New("cannot parse 'created_at' field")
		return result, err
	}

	return
}
