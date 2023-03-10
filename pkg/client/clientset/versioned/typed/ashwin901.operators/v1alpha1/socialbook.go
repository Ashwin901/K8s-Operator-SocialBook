/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/ashwin901/social-book-operator/pkg/apis/ashwin901.operators/v1alpha1"
	scheme "github.com/ashwin901/social-book-operator/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// SocialBooksGetter has a method to return a SocialBookInterface.
// A group's client should implement this interface.
type SocialBooksGetter interface {
	SocialBooks(namespace string) SocialBookInterface
}

// SocialBookInterface has methods to work with SocialBook resources.
type SocialBookInterface interface {
	Create(ctx context.Context, socialBook *v1alpha1.SocialBook, opts v1.CreateOptions) (*v1alpha1.SocialBook, error)
	Update(ctx context.Context, socialBook *v1alpha1.SocialBook, opts v1.UpdateOptions) (*v1alpha1.SocialBook, error)
	UpdateStatus(ctx context.Context, socialBook *v1alpha1.SocialBook, opts v1.UpdateOptions) (*v1alpha1.SocialBook, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.SocialBook, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.SocialBookList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.SocialBook, err error)
	SocialBookExpansion
}

// socialBooks implements SocialBookInterface
type socialBooks struct {
	client rest.Interface
	ns     string
}

// newSocialBooks returns a SocialBooks
func newSocialBooks(c *OperatorsV1alpha1Client, namespace string) *socialBooks {
	return &socialBooks{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the socialBook, and returns the corresponding socialBook object, and an error if there is any.
func (c *socialBooks) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.SocialBook, err error) {
	result = &v1alpha1.SocialBook{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("socialbooks").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SocialBooks that match those selectors.
func (c *socialBooks) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.SocialBookList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.SocialBookList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("socialbooks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested socialBooks.
func (c *socialBooks) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("socialbooks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a socialBook and creates it.  Returns the server's representation of the socialBook, and an error, if there is any.
func (c *socialBooks) Create(ctx context.Context, socialBook *v1alpha1.SocialBook, opts v1.CreateOptions) (result *v1alpha1.SocialBook, err error) {
	result = &v1alpha1.SocialBook{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("socialbooks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(socialBook).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a socialBook and updates it. Returns the server's representation of the socialBook, and an error, if there is any.
func (c *socialBooks) Update(ctx context.Context, socialBook *v1alpha1.SocialBook, opts v1.UpdateOptions) (result *v1alpha1.SocialBook, err error) {
	result = &v1alpha1.SocialBook{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("socialbooks").
		Name(socialBook.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(socialBook).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *socialBooks) UpdateStatus(ctx context.Context, socialBook *v1alpha1.SocialBook, opts v1.UpdateOptions) (result *v1alpha1.SocialBook, err error) {
	result = &v1alpha1.SocialBook{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("socialbooks").
		Name(socialBook.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(socialBook).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the socialBook and deletes it. Returns an error if one occurs.
func (c *socialBooks) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("socialbooks").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *socialBooks) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("socialbooks").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched socialBook.
func (c *socialBooks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.SocialBook, err error) {
	result = &v1alpha1.SocialBook{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("socialbooks").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
