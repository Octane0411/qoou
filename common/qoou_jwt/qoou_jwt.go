package qoou_jwt

import (
	"crypto/rsa"
	"github.com/golang-jwt/jwt"
	"time"
)

var VerifyKey *rsa.PublicKey
var SignKey *rsa.PrivateKey

func init() {
	VerifyKey, _ = jwt.ParseRSAPublicKeyFromPEM([]byte(`
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCrhAnBMfs201KAiSTyeSjzre9m
2518EhLRl3EBfjRvUr7iETxOp0qKHittFX7bxLHLl74nKCC7lXv9fwKanaGXf0fr
yZLYCKJa5ve81gh0WYJ24smADwfx78smWCYmJuJ62kn2Lbr4xbFvpEqLFBynfK+Z
iCi/yHJPD4rvu7OPpQIDAQAB
-----END PUBLIC KEY-----
`))
	SignKey, _ = jwt.ParseRSAPrivateKeyFromPEM([]byte(`
-----BEGIN PRIVATE KEY-----
MIICdwIBADANBgkqhkiG9w0BAQEFAASCAmEwggJdAgEAAoGBAKuECcEx+zbTUoCJ
JPJ5KPOt72bbnXwSEtGXcQF+NG9SvuIRPE6nSooeK20VftvEscuXvicoILuVe/1/
ApqdoZd/R+vJktgIolrm97zWCHRZgnbiyYAPB/HvyyZYJiYm4nraSfYtuvjFsW+k
SosUHKd8r5mIKL/Ick8Piu+7s4+lAgMBAAECgYEAgI/EUBAK4ZmdKcOi8i1nSOCD
pnHPpgRWHsyJZDkZTKiVdBa/QaWb9dOPcYC/SjQxoQ3o9qjZgEIYYnclmIe3awFG
BIvluIj/ZAoVm/CRtPT0A4be5halixJ2BKIgGAY0Vnzrh852PP9rupL2MNrKj2tI
26RBtbneBSiqR7LwToECQQDV1bVvX0I6DAUoEDnCqyl7P+wp+zOSW0ZyDIiZIEWF
tCrVQNYYujYh2F+xN08rBP8dMvMU+71v/FEroLSBdrs5AkEAzVYVKCPtnRniNsF1
fnq3ypkMB9FA+O9IP0QNFGaxNzO1N3MekiBH9RPwiuhbPFnpwTqI6AxAQg26podc
UsW7zQJAd8PEZOZzj1NgJ/o+f5uiFhfNTA4X6mcY45PFhg4fIi2wt9QilaLl4rrv
jbAutSeNQ2tf3mbIyUoGpGrT7pbzcQJADmf5uAU9SIZmXp0YFzWY63ftZicCPfTb
xsSJfmLuEAdqsWc8P9hP9BvgBn7i18sfIVVwAYfKglfgPorEqXICCQJBAKeyzIdO
V2u4qdHk5MtviKQnXDgDgo73HAhfqPcgHY9XnZuAS/jFv5XyPPurC0ZvpeSST4p7
yoAodhzgDWtzKRM=
-----END PRIVATE KEY-----
`))

}

type QoouCliams struct {
	Username       string    `json:"username"`
	ExpirationTime time.Time `json:"expirationTime"`
}

func (q QoouCliams) Valid() error {
	return nil
}

func CreateToken(username string) (string, error) {
	t := jwt.New(jwt.GetSigningMethod("RS256"))

	t.Claims = &QoouCliams{
		Username:       username,
		ExpirationTime: time.Now().Add(time.Minute * time.Duration(1)),
	}

	return t.SignedString(SignKey)
}
