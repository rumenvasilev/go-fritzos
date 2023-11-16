package auth

import (
	"testing"

	"github.com/matryer/is"
)

func Test_solveChallenge(t *testing.T) {
	is := is.New(t)

	t.Run("v1 challenge", func(t *testing.T) {
		// MD5, v1 challenge
		challenge := "1234567z"
		password := "äbc"
		want := "1234567z-9e224a41eeefa284df7bb0f26c2913e2"

		got, err := solveChallenge(challenge, password)
		is.NoErr(err)
		is.Equal(got, want)
	})

	t.Run("v2 challenge", func(t *testing.T) {
		// V2 challenge
		challenge := "2$10000$5A1711$2000$5A1722"
		password := "1example!"
		want := "5A1722$1798a1672bca7c6463d6b245f82b53703b0f50813401b03e4045a5861e689adb"

		got, err := solveChallenge(challenge, password)
		is.NoErr(err)
		is.Equal(got, want)
	})
}
