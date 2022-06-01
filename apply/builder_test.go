package apply

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	testFile1     = "/dir/file1.yaml"
	testFile2     = "/dir/file2.yaml"
	testNamespace = "le-namespace"
)

//go:embed testdata/multi-doc.yaml
var multiDocYamlBytes []byte

//go:embed testdata/multi-doc-template.yaml
var multiDocYamlTemplateBytes []byte

func TestBuilder_WithYamlResource(t *testing.T) {
	t.Run("should add a single resource", func(t *testing.T) {
		sut := &Builder{
			fileToGenericResource: make(map[string][]byte),
			fileToTemplate:        make(map[string]interface{}),
		}

		// when
		sut.WithYamlResource(testFile1, multiDocYamlBytes)

		// then
		assert.NotEmpty(t, sut.fileToGenericResource[testFile1])
		assert.Equal(t, multiDocYamlBytes, sut.fileToGenericResource[testFile1])
	})
	t.Run("should distinguish between different files", func(t *testing.T) {
		sut := &Builder{
			fileToGenericResource: make(map[string][]byte),
			fileToTemplate:        make(map[string]interface{}),
		}

		// when
		sut.WithYamlResource(testFile1, multiDocYamlBytes).
			WithYamlResource(testFile2, multiDocYamlTemplateBytes)

		// then
		require.Len(t, sut.fileToGenericResource, 2)

		assert.NotEmpty(t, sut.fileToGenericResource[testFile1])
		assert.Equal(t, multiDocYamlBytes, sut.fileToGenericResource[testFile1])

		assert.NotEmpty(t, sut.fileToGenericResource[testFile2])
		assert.Equal(t, multiDocYamlTemplateBytes, sut.fileToGenericResource[testFile2])
	})
}

func TestBuilder_WithTemplate(t *testing.T) {
	t.Run("should add a single template", func(t *testing.T) {
		sut := &Builder{
			fileToGenericResource: make(map[string][]byte),
			fileToTemplate:        make(map[string]interface{}),
		}
		templateObj := struct {
			Namespace string
		}{
			Namespace: testNamespace,
		}

		// when
		sut.WithTemplate(testFile2, templateObj)

		// then
		assert.NotEmpty(t, sut.fileToTemplate[testFile2])
		assert.Equal(t, templateObj, sut.fileToTemplate[testFile2])
	})
	t.Run("should maintain two different template objects", func(t *testing.T) {
		sut := &Builder{
			fileToGenericResource: make(map[string][]byte),
			fileToTemplate:        make(map[string]interface{}),
		}
		templateObj1 := struct {
			Namespace string
		}{
			Namespace: testNamespace,
		}
		templateObj2 := struct {
			Namespace string
		}{
			Namespace: "hello-world",
		}

		// when
		sut.WithTemplate(testFile1, templateObj1).
			WithTemplate(testFile2, templateObj2)

		// then
		require.Len(t, sut.fileToTemplate, 2)
		assert.NotEmpty(t, sut.fileToTemplate[testFile1])
		assert.Equal(t, templateObj1, sut.fileToTemplate[testFile1])

		assert.NotEmpty(t, sut.fileToTemplate[testFile2])
		assert.Equal(t, templateObj2, sut.fileToTemplate[testFile2])
	})
}

func TestBuilder_WithOwner(t *testing.T) {
	t.Run("should add an owner resource for all generic resources", func(t *testing.T) {
		sut := &Builder{
			fileToGenericResource: make(map[string][]byte),
			fileToTemplate:        make(map[string]interface{}),
		}
		anyObject := &v1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "le-service-account",
				Namespace: testNamespace,
			},
		}

		// when
		sut.WithOwner(anyObject)

		// then
		assert.NotNil(t, sut.owningResource)
		assert.Equal(t, anyObject, sut.owningResource)
	})
}

func Test_renderTemplate(t *testing.T) {
	t.Run("should template namespace", func(t *testing.T) {
		tempDoc := []byte(`hello {{ .Namespace }}`)
		templateObj1 := struct {
			Namespace string
		}{
			Namespace: testNamespace,
		}

		actual, err := renderTemplate(testFile1, tempDoc, templateObj1)

		require.NoError(t, err)
		expected := []byte(`hello le-namespace`)
		assert.Equal(t, expected, actual)
	})

	t.Run("should return error", func(t *testing.T) {
		tempDoc := []byte(`hello {{ .Namespace `)
		templateObj1 := struct {
			Namespace string
		}{
			Namespace: testNamespace,
		}

		_, err := renderTemplate(testFile1, tempDoc, templateObj1)

		require.Error(t, err)
		assert.Equal(t, "failed to parse template for file /dir/file1.yaml: template: t:1: unclosed action", err.Error())
	})
}

type mockApplier struct {
	mock.Mock
}

func (m *mockApplier) ApplyWithOwner(doc YamlDocument, namespace string, resource metav1.Object) error {
	args := m.Called(doc, namespace, resource)
	return args.Error(0)
}
