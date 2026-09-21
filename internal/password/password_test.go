package password

import "testing"

func TestSaltedHashAndVerification(t *testing.T) {
	secret := "a long personal passphrase"
	a, err := Hash(secret)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Hash(secret)
	if err != nil || a == b {
		t.Fatal("hashes must have unique salts", err)
	}
	for _, v := range []struct {
		value, hash string
		want        bool
	}{{secret, a, true}, {"wrong", a, false}, {"wrong", DummyHash, false}} {
		got, err := Verify(v.value, v.hash)
		if err != nil || got != v.want {
			t.Fatal("verification failed", err)
		}
	}
	if _, err = Hash("short"); err == nil {
		t.Fatal("weak length accepted")
	}
	if _, err = Verify(secret, "$argon2id$v=19$m=999999999,t=2,p=1$x$x"); err == nil {
		t.Fatal("unbounded parameters accepted")
	}
}
