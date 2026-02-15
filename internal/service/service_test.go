package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"green-api/internal/greenapi"
)

type mockClient struct {
	sendFileByURLFn func(ctx context.Context, idInstance, apiTokenInstance, chatID, urlFile, fileName string) (greenapi.Response, error)
}

func (m *mockClient) GetSettings(context.Context, string, string) (greenapi.Response, error) {
	return greenapi.Response{}, nil
}

func (m *mockClient) GetStateInstance(context.Context, string, string) (greenapi.Response, error) {
	return greenapi.Response{}, nil
}

func (m *mockClient) SendMessage(context.Context, string, string, string, string) (greenapi.Response, error) {
	return greenapi.Response{}, nil
}

func (m *mockClient) SendFileByURL(ctx context.Context, idInstance, apiTokenInstance, chatID, urlFile, fileName string) (greenapi.Response, error) {
	if m.sendFileByURLFn == nil {
		return greenapi.Response{}, nil
	}
	return m.sendFileByURLFn(ctx, idInstance, apiTokenInstance, chatID, urlFile, fileName)
}

func TestNormalizeChatID_AddSuffix(t *testing.T) {
	t.Parallel()

	chatID, err := NormalizeChatID("77771234567")
	require.NoError(t, err)
	require.Equal(t, "77771234567@c.us", chatID)
}

func TestNormalizeChatID_PreserveSuffix(t *testing.T) {
	t.Parallel()

	chatID, err := NormalizeChatID("77771234567@c.us")
	require.NoError(t, err)
	require.Equal(t, "77771234567@c.us", chatID)
}

func TestNormalizeChatID_Invalid(t *testing.T) {
	t.Parallel()

	_, err := NormalizeChatID("abc")
	require.Error(t, err)
}

func TestExtractFileName(t *testing.T) {
	t.Parallel()

	fileName, err := ExtractFileName("https://example.com/path/horse.png")
	require.NoError(t, err)
	require.Equal(t, "horse.png", fileName)
}

func TestSendFileByURL_ComputesFileNameAndNormalizesChatID(t *testing.T) {
	t.Parallel()

	called := false
	client := &mockClient{
		sendFileByURLFn: func(_ context.Context, idInstance, apiTokenInstance, chatID, urlFile, fileName string) (greenapi.Response, error) {
			called = true
			require.Equal(t, "1101000001", idInstance)
			require.Equal(t, "token", apiTokenInstance)
			require.Equal(t, "77771234567@c.us", chatID)
			require.Equal(t, "https://my.site.com/img/horse.png", urlFile)
			require.Equal(t, "horse.png", fileName)
			return greenapi.Response{StatusCode: http.StatusOK, Body: []byte(`{"ok":true}`), ContentType: "application/json"}, nil
		},
	}

	svc := New(client)
	resp, apiErr := svc.SendFileByURL(context.Background(), SendFileByURLRequest{
		CredentialsRequest: CredentialsRequest{
			IDInstance:       "1101000001",
			APITokenInstance: "token",
		},
		ChatID:  "77771234567",
		URLFile: "https://my.site.com/img/horse.png",
	})

	require.Nil(t, apiErr)
	require.True(t, called)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
