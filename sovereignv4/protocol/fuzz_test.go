package protocol

import (
	"bytes"
	"testing"
)

func FuzzCertificateCodec(f *testing.F) {
	fx := makeFixtures(f)
	seed, _ := fx.certificate.MarshalBinary()
	f.Add(seed)
	f.Fuzz(func(t *testing.T, b []byte) {
		v, err := UnmarshalCertificate(b)
		if err != nil {
			return
		}
		canonical, err := v.MarshalBinary()
		if err != nil || !bytes.Equal(canonical, b) {
			t.Fatal("successful decode was not canonical")
		}
	})
}

func FuzzEnvelopeCodec(f *testing.F) {
	fx := makeFixtures(f)
	seed, _ := fx.envelope.MarshalBinary()
	f.Add(seed)
	f.Fuzz(func(t *testing.T, b []byte) {
		v, err := UnmarshalEnvelope(b)
		if err != nil {
			return
		}
		canonical, err := v.MarshalBinary()
		if err != nil || !bytes.Equal(canonical, b) {
			t.Fatal("successful decode was not canonical")
		}
	})
}

func FuzzManifestCodec(f *testing.F) {
	fx := makeFixtures(f)
	seed, _ := fx.manifest.MarshalBinary()
	f.Add(seed)
	f.Fuzz(func(t *testing.T, b []byte) {
		v, err := UnmarshalManifest(b)
		if err != nil {
			return
		}
		canonical, err := v.MarshalBinary()
		if err != nil || !bytes.Equal(canonical, b) {
			t.Fatal("successful decode was not canonical")
		}
	})
}
