package utils

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

const smallEnglishLettersFixture = "abcdefghijklmnopqrstuvwxyz"
const smallEnglishLettersSFixture = "^"
const eightaFixture = "aaaaaaaa"

type PasswordGeneratorTestSuite struct {
	suite.Suite
}

func TestNewPasswordResetRepository(t *testing.T) {
	s := new(PasswordGeneratorTestSuite)
	suite.Run(t, s)
}

func (s *PasswordGeneratorTestSuite) TestPasswordGenerator_generateRandomResetPassword_gen_8_symbols_a_no_error() {
	t := s.T()

	res, err := generateRandomResetPassword(8, "a")
	assert.Equal(t, nil, err)
	assert.Equal(t, eightaFixture, res)
}

func (s *PasswordGeneratorTestSuite) TestPasswordGenerator_generateRandomResetPassword_len_8_symbols_any_no_error() {
	t := s.T()

	ret, err := generateRandomResetPassword(8, smallEnglishLettersFixture)
	assert.Equal(t, nil, err)
	assert.Equal(t, 8, len(ret))

	for _, ch := range ret {
		if !strings.Contains(smallEnglishLettersFixture, string(ch)) {
			t.Errorf("unsupported symbol %c in generated string", ch)
			break
		}
	}

}

func (s *PasswordGeneratorTestSuite) TestPasswordGenerator_generateRandomResetPassword_len_out_and_on_border_of_interval_symbols_any() {
	t := s.T()

	_, err := generateRandomResetPassword(-8, smallEnglishLettersFixture)
	assert.Error(t, err)

	_, err = generateRandomResetPassword(MinPassLen, smallEnglishLettersFixture)
	assert.NoError(t, err)

	_, err = generateRandomResetPassword(MinPassLen, smallEnglishLettersFixture)
	assert.NoError(t, err)
}
