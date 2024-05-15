package email

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

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
		require.Equal(t, conf.SenderFromName, d.FromName)
		require.Equal(t, conf.SenderFromAddress, d.FromAddr)
		require.Equal(t, email, d.ToAddr)
		require.Equal(t, "Password Reset", d.Subject)
		return true
	})).Return(nil)

	s := NewSenderSmtp(conf, cl)
	require.NoError(t, s.SendResetLink(email, name, "123"))

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
		require.Equal(t, conf.SenderFromName, d.FromName)
		require.Equal(t, conf.SenderFromAddress, d.FromAddr)
		require.Equal(t, email, d.ToAddr)
		require.Equal(t, "New Password", d.Subject)
		return true
	})).Return(nil)

	s := NewSenderSmtp(conf, cl)
	require.NoError(t, s.SendNewPassword(email, name, "123"))

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
		require.Equal(t, conf.SenderFromName, d.FromName)
		require.Equal(t, conf.SenderFromAddress, d.FromAddr)
		require.Equal(t, email, d.ToAddr)
		require.Equal(t, "Registration confirmation", d.Subject)
		return true
	})).Return(nil)

	s := NewSenderSmtp(conf, cl)
	require.NoError(t, s.SendRegistrationConfirmLink(email, name, "123"))

	cl.AssertExpectations(t)
}

func TestSenderImpl_SendEmailConfirmationLink(t *testing.T) {
	name := "1"
	email := "2"

	conf := config.Email{
		SenderFromName:    "sender_from_name",
		SenderFromAddress: "sender_from_name@gmail.com",
		IsSendRequired:    true,
	}

	cl := mocks.NewSMTPClient(t)
	cl.On("Send", mock.MatchedBy(func(d *domain.SendData) bool {
		require.Equal(t, conf.SenderFromName, d.FromName)
		require.Equal(t, conf.SenderFromAddress, d.FromAddr)
		require.Equal(t, email, d.ToAddr)
		require.Equal(t, "Email confirmation", d.Subject)
		return true
	})).Return(nil)

	s := NewSenderSmtp(conf, cl)
	require.NoError(t, s.SendEmailConfirmationLink(email, name, "123"))

	cl.AssertExpectations(t)
}
