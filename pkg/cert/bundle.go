package cert

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type Serving struct {
	ServiceKey    []byte
	ServiceCert   []byte
	ServingCertCA []byte
}

type Bundle struct {
	Serving
	KubeClintCA []byte
}

func (b *Bundle) Validate() error {
	var (
		EmptyKubeClintCAErr          = errors.New("kube client CA must be specified")
		EmptyServingServiceCertErr   = errors.New("serving service cert must be specified")
		EmptyServingServiceKeyErr    = errors.New("serving service private key must be specified")
		EmptyServingServiceCertCAErr = errors.New("serving service cert CA must be specified")
	)

	if len(b.KubeClintCA) == 0 {
		return EmptyKubeClintCAErr
	}

	if len(b.ServingCertCA) == 0 {
		return EmptyServingServiceCertCAErr
	}

	if len(b.Serving.ServiceCert) == 0 {
		return EmptyServingServiceCertErr
	}

	if len(b.Serving.ServiceKey) == 0 {
		return EmptyServingServiceKeyErr
	}

	return nil
}

func (s *Serving) Hash() string {
	writer := sha256.New()

	writer.Write(s.ServiceKey)
	h1 := writer.Sum(nil)

	writer.Reset()
	writer.Write(s.ServiceCert)
	h2 := writer.Sum(h1)

	writer.Reset()
	writer.Write(s.ServingCertCA)
	h := writer.Sum(h2)

	writer.Reset()
	writer.Write(h)

	return hex.EncodeToString(writer.Sum(nil))
}
