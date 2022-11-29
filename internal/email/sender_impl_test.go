package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/config"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/internal/generated/mocks"
	"git.epam.com/epm-lstr/epm-lstr-lc/be/pkg/domain"
)

func TestSenderImpl_SendResetLink(t *testing.T) {
	name := "1"
	email := "2"

	conf := config.Email{
		SenderFromName:    "sender_from_name",
		SenderFromAddress: "sender_from_name@gmail.com",
		IsSendRequired:    true,
	}

	cl := mocks.NewSMTPClient(t)
	cl.On("Send", mock.MatchedBy(func(d *domain.SendData) bool {
		assert.Equal(t, conf.SenderFromName, d.FromName)
		assert.Equal(t, conf.SenderFromAddress, d.FromAddr)
		assert.Equal(t, email, d.ToAddr)
		assert.Equal(t, "Password Reset", d.Subject)
		return true
	})).Return(nil)

	s := NewSenderSmtp(conf, cl)
	assert.NoError(t, s.SendResetLink(email, name, "123"))

	cl.AssertExpectations(t)
}

func TestSenderImpl_SendNewPassword(t *testing.T) {
	name := "1"
	email := "2"

	conf := config.Email{
		SenderFromName:    "sender_from_name",
		SenderFromAddress: "sender_from_name@gmail.com",
		IsSendRequired:    true,
	}

	cl := mocks.NewSMTPClient(t)
	cl.On("Send", mock.MatchedBy(func(d *domain.SendData) bool {
		assert.Equal(t, conf.SenderFromName, d.FromName)
		assert.Equal(t, conf.SenderFromAddress, d.FromAddr)
		assert.Equal(t, email, d.ToAddr)
		assert.Equal(t, "New Password", d.Subject)
		return true
	})).Return(nil)

	s := NewSenderSmtp(conf, cl)
	assert.NoError(t, s.SendNewPassword(email, name, "123"))

	cl.AssertExpectations(t)
}

func TestSenderImpl_SendRegistrationConfirmLink(t *testing.T) {
	name := "1"
	email := "2"

	conf := config.Email{
		SenderFromName:    "sender_from_name",
		SenderFromAddress: "sender_from_name@gmail.com",
		IsSendRequired:    true,
	}

	cl := mocks.NewSMTPClient(t)
	cl.On("Send", mock.MatchedBy(func(d *domain.SendData) bool {
		assert.Equal(t, conf.SenderFromName, d.FromName)
		assert.Equal(t, conf.SenderFromAddress, d.FromAddr)
		assert.Equal(t, email, d.ToAddr)
		assert.Equal(t, "Registration confirmation", d.Subject)
		return true
	})).Return(nil)

	s := NewSenderSmtp(conf, cl)
	assert.NoError(t, s.SendRegistrationConfirmLink(email, name, "123"))

	cl.AssertExpectations(t)
}
