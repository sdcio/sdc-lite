package schemaloader

import (
	"context"

	"github.com/sdcio/config-server/pkg/git/auth"
	"k8s.io/apimachinery/pkg/types"
)

func NewNopResolver() auth.CredentialResolver {
	return &secretNopResolver{}
}

var _ auth.CredentialResolver = &secretNopResolver{}

type secretNopResolver struct{}

func (r *secretNopResolver) ResolveCredential(ctx context.Context, nsn types.NamespacedName) (auth.Credential, error) {
	return nil, nil
}
