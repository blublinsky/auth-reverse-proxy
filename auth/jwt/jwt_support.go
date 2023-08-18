package jwt

import (
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"k8s.io/klog/v2"
)

const(
	claim_duration = 5
	cookie_name = "token"
	token_header = "auth-token"
	auth_token = "12345"
)

var jwtKey = []byte("my_secret_key")

// Create a struct that will be encoded to a JWT.
// We add jwt.RegisteredClaims as an embedded type, to provide fields like expiry time
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Create a new JWT token string
func NewJWTTokenString(user string) *string {
	// Declare the expiration time of the token
	// here, we have kept it as 5 minutes
	expirationTime := time.Now().Add(claim_duration * time.Minute)
	// Create the JWT claims, which includes the username and expiry time
	claims := &Claims{
		Username: user,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		klog.Info("Failed to create JWT token, err ", err)
		return nil
	}
	return &tokenString
}

// Validate token string
// Returns: 
//		0 - token is valid
//		1 - unauthorised
//		2 - bad request
func ValidateTokenString(tokenString string) int {
	// Initialize a new instance of `Claims`
	claims := &Claims{}

	// Parse the JWT string and store the result in `claims`.
	// Note that we are passing the key in this method as well. This method will return an error
	// if the token is invalid (if it has expired according to the expiry time we set on sign in),
	// or if the signature does not match
	tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			// Invalid signature
			klog.Info("Invalid signature parsing token")
			return 1
		}
		// Other errors
		klog.Info("Error parsing token, err ", err)
		return 2
	}
	if !tkn.Valid {
		klog.Info("Got invalid token")
		return 1
	}
	return 0
}

// Set user cookie with a new token
func SetTokenCookie(writer http.ResponseWriter, tkn *string, expry *time.Time ){
	http.SetCookie(writer, &http.Cookie{
		Name:    cookie_name,
		Value:   *tkn,
		Expires: *expry,
	})
}

// Get JWT token from cookie. Returns the following error codes:
//		0 - success
//		1 - unauthorised
//		2 - bad request
func GetTokenFromCookie(req *http.Request) (*string, int) {
		// We can obtain the session token from the requests cookies, which come with every request
		c, err := req.Cookie(cookie_name)
		if err != nil {
			if err == http.ErrNoCookie {
				// If the cookie is not set, return an unauthorized status
				return nil, 1
			}
			// For any other type of error, return a bad request status
			return nil, 2
		}
	
		// Get the JWT string from the cookie
		return &c.Value, 0	
}

// Validate user
// This is a very simple user validation based on a user token
// In a real implementation this needs to be replaced with the user validation based on
// the external user management system, for example LDAP or KeyCloak
func ValidateUser(req *http.Request) *string {
	current := req.Header.Get(token_header)
	klog.Info("Auth token ", current)
	if current != auth_token {
		// Wrong token
		return nil
	}
	return &current
}