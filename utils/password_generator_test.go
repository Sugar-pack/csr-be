package utils

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	smallEnglishLettersFixture = "abcdefghijklmnopqrstuvwxyz"
	eightAsFixture             = "aaaaaaaa"
	invalidPasswordLenFixture  = -8
	validPasswordLenFixture    = 8
)

type PasswordGeneratorTestSuite struct {
	suite.Suite
}

func TestNewPasswordGenerator(t *testing.T) {
	s := new(PasswordGeneratorTestSuite)
	suite.Run(t, s)
}

func (s *PasswordGeneratorTestSuite) TestPasswordGenerator_generateRandomPassword_gen_8_symbols_a_no_error() {
	t := s.T()

	res, err := generateRandomPassword(validPasswordLenFixture, "a")
	assert.Equal(t, nil, err)
	assert.Equal(t, eightAsFixture, res)
}

func (s *PasswordGeneratorTestSuite) TestPasswordGenerator_generateRandomPassword_len_8_symbols_any_no_error() {
	t := s.T()

	ret, err := generateRandomPassword(validPasswordLenFixture, smallEnglishLettersFixture)
	assert.Equal(t, nil, err)
	assert.Equal(t, validPasswordLenFixture, len(ret))
	for _, ch := range ret {
		if !strings.Contains(smallEnglishLettersFixture, string(ch)) {
			t.Errorf("unsupported symbol %c in generated string", ch)
			break
		}
	}

}

func (s *PasswordGeneratorTestSuite) TestPasswordGenerator_generateRandomPassword_len_out_and_on_border_of_interval_symbols_any() {
	t := s.T()

	_, err := generateRandomPassword(invalidPasswordLenFixture, smallEnglishLettersFixture)
	assert.Error(t, err)

	_, err = generateRandomPassword(MinPasswordLen, smallEnglishLettersFixture)
	assert.NoError(t, err)

	_, err = generateRandomPassword(MaxPasswordLen, smallEnglishLettersFixture)

	assert.NoError(t, err)
}
