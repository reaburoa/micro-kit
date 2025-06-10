package env

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_findDir(t *testing.T) {
	pwd, _ := os.Getwd()

	parentPath := findDir(pwd, "ctxutils")
	require.True(t, strings.HasSuffix(parentPath, "/common/utils"))
	parentPath = findDir(pwd, "ctxutilsdcgsiwdvgweodhwefdpwjcd")
	require.Equal(t, parentPath, "")

}
