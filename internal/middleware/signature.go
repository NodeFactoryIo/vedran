package middleware

import (
	"github.com/NodeFactoryIo/go-substrate-rpc-client/signature"
	"github.com/NodeFactoryIo/vedran/internal/constants"
	"github.com/ethereum/go-ethereum/common/hexutil"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func VerifySignatureMiddleware(next http.Handler, privateKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		verified, httpStatusCode, err := verifySignatureInHeader(r, privateKey)
		if err != nil {
			http.Error(w, http.StatusText(httpStatusCode), httpStatusCode)
			return
		}
		if !verified {
			log.Errorf("Invalid request signature")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func verifySignatureInHeader(r *http.Request, privateKey string) (bool, int, error) {
	sig := r.Header.Get("X-Signature")
	if sig == "" {
		log.Error("Missing signature header")
		return false, http.StatusBadRequest, nil
	}
	sigInBytes, err := hexutil.Decode(sig)
	if err != nil {
		log.Errorf("Unable to decode signature, because of: %v", err)
		return false, http.StatusBadRequest, err
	}
	verified, err := signature.Verify([]byte(constants.StatsSignedData), sigInBytes, privateKey)
	if err != nil {
		log.Errorf("Failed to verify signature, because %v", err)
		return false, http.StatusInternalServerError, err
	}
	return verified, 0, err
}
