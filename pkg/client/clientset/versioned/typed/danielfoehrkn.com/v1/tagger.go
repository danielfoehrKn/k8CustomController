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

package v1

import (
	v1 "github.com/danielfoehrkn/custom-database-controller/pkg/apis/danielfoehrkn.com/v1"
	scheme "github.com/danielfoehrkn/custom-database-controller/pkg/client/clientset/versioned/scheme"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// TaggersGetter has a method to return a TaggerInterface.
// A group's client should implement this interface.
type TaggersGetter interface {
	Taggers(namespace string) TaggerInterface
}

// TaggerInterface has methods to work with Tagger resources.
type TaggerInterface interface {
	Create(*v1.Tagger) (*v1.Tagger, error)
	Update(*v1.Tagger) (*v1.Tagger, error)
	UpdateStatus(*v1.Tagger) (*v1.Tagger, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	Get(name string, options metav1.GetOptions) (*v1.Tagger, error)
	List(opts metav1.ListOptions) (*v1.TaggerList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Tagger, err error)
	GetScale(taggerName string, options metav1.GetOptions) (*autoscalingv1.Scale, error)
	UpdateScale(taggerName string, scale *autoscalingv1.Scale) (*autoscalingv1.Scale, error)

	TaggerExpansion
}

// taggers implements TaggerInterface
type taggers struct {
	client rest.Interface
	ns     string
}

// newTaggers returns a Taggers
func newTaggers(c *DanielfoehrknV1Client, namespace string) *taggers {
	return &taggers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the tagger, and returns the corresponding tagger object, and an error if there is any.
func (c *taggers) Get(name string, options metav1.GetOptions) (result *v1.Tagger, err error) {
	result = &v1.Tagger{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("taggers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Taggers that match those selectors.
func (c *taggers) List(opts metav1.ListOptions) (result *v1.TaggerList, err error) {
	result = &v1.TaggerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("taggers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested taggers.
func (c *taggers) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("taggers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a tagger and creates it.  Returns the server's representation of the tagger, and an error, if there is any.
func (c *taggers) Create(tagger *v1.Tagger) (result *v1.Tagger, err error) {
	result = &v1.Tagger{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("taggers").
		Body(tagger).
		Do().
		Into(result)
	return
}

// Update takes the representation of a tagger and updates it. Returns the server's representation of the tagger, and an error, if there is any.
func (c *taggers) Update(tagger *v1.Tagger) (result *v1.Tagger, err error) {
	result = &v1.Tagger{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("taggers").
		Name(tagger.Name).
		Body(tagger).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *taggers) UpdateStatus(tagger *v1.Tagger) (result *v1.Tagger, err error) {
	result = &v1.Tagger{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("taggers").
		Name(tagger.Name).
		SubResource("status").
		Body(tagger).
		Do().
		Into(result)
	return
}

// Delete takes name of the tagger and deletes it. Returns an error if one occurs.
func (c *taggers) Delete(name string, options *metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("taggers").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *taggers) DeleteCollection(options *metav1.DeleteOptions, listOptions metav1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("taggers").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched tagger.
func (c *taggers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.Tagger, err error) {
	result = &v1.Tagger{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("taggers").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}

// GetScale takes name of the tagger, and returns the corresponding autoscalingv1.Scale object, and an error if there is any.
func (c *taggers) GetScale(taggerName string, options metav1.GetOptions) (result *autoscalingv1.Scale, err error) {
	result = &autoscalingv1.Scale{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("taggers").
		Name(taggerName).
		SubResource("scale").
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// UpdateScale takes the top resource name and the representation of a scale and updates it. Returns the server's representation of the scale, and an error, if there is any.
func (c *taggers) UpdateScale(taggerName string, scale *autoscalingv1.Scale) (result *autoscalingv1.Scale, err error) {
	result = &autoscalingv1.Scale{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("taggers").
		Name(taggerName).
		SubResource("scale").
		Body(scale).
		Do().
		Into(result)
	return
}