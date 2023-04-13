// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import (
	context "context"

	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mock "github.com/stretchr/testify/mock"

	types "k8s.io/apimachinery/pkg/types"

	v1 "k8s.io/client-go/applyconfigurations/core/v1"

	watch "k8s.io/apimachinery/pkg/watch"
)

// ConfigMaps is an autogenerated mock type for the ConfigMaps type
type ConfigMaps struct {
	mock.Mock
}

type ConfigMaps_Expecter struct {
	mock *mock.Mock
}

func (_m *ConfigMaps) EXPECT() *ConfigMaps_Expecter {
	return &ConfigMaps_Expecter{mock: &_m.Mock}
}

// Apply provides a mock function with given fields: ctx, configMap, opts
func (_m *ConfigMaps) Apply(ctx context.Context, configMap *v1.ConfigMapApplyConfiguration, opts metav1.ApplyOptions) (*corev1.ConfigMap, error) {
	ret := _m.Called(ctx, configMap, opts)

	var r0 *corev1.ConfigMap
	if rf, ok := ret.Get(0).(func(context.Context, *v1.ConfigMapApplyConfiguration, metav1.ApplyOptions) *corev1.ConfigMap); ok {
		r0 = rf(ctx, configMap, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *v1.ConfigMapApplyConfiguration, metav1.ApplyOptions) error); ok {
		r1 = rf(ctx, configMap, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConfigMaps_Apply_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Apply'
type ConfigMaps_Apply_Call struct {
	*mock.Call
}

// Apply is a helper method to define mock.On call
//   - ctx context.Context
//   - configMap *v1.ConfigMapApplyConfiguration
//   - opts metav1.ApplyOptions
func (_e *ConfigMaps_Expecter) Apply(ctx interface{}, configMap interface{}, opts interface{}) *ConfigMaps_Apply_Call {
	return &ConfigMaps_Apply_Call{Call: _e.mock.On("Apply", ctx, configMap, opts)}
}

func (_c *ConfigMaps_Apply_Call) Run(run func(ctx context.Context, configMap *v1.ConfigMapApplyConfiguration, opts metav1.ApplyOptions)) *ConfigMaps_Apply_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v1.ConfigMapApplyConfiguration), args[2].(metav1.ApplyOptions))
	})
	return _c
}

func (_c *ConfigMaps_Apply_Call) Return(result *corev1.ConfigMap, err error) *ConfigMaps_Apply_Call {
	_c.Call.Return(result, err)
	return _c
}

// Create provides a mock function with given fields: ctx, configMap, opts
func (_m *ConfigMaps) Create(ctx context.Context, configMap *corev1.ConfigMap, opts metav1.CreateOptions) (*corev1.ConfigMap, error) {
	ret := _m.Called(ctx, configMap, opts)

	var r0 *corev1.ConfigMap
	if rf, ok := ret.Get(0).(func(context.Context, *corev1.ConfigMap, metav1.CreateOptions) *corev1.ConfigMap); ok {
		r0 = rf(ctx, configMap, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *corev1.ConfigMap, metav1.CreateOptions) error); ok {
		r1 = rf(ctx, configMap, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConfigMaps_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type ConfigMaps_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - configMap *corev1.ConfigMap
//   - opts metav1.CreateOptions
func (_e *ConfigMaps_Expecter) Create(ctx interface{}, configMap interface{}, opts interface{}) *ConfigMaps_Create_Call {
	return &ConfigMaps_Create_Call{Call: _e.mock.On("Create", ctx, configMap, opts)}
}

func (_c *ConfigMaps_Create_Call) Run(run func(ctx context.Context, configMap *corev1.ConfigMap, opts metav1.CreateOptions)) *ConfigMaps_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*corev1.ConfigMap), args[2].(metav1.CreateOptions))
	})
	return _c
}

func (_c *ConfigMaps_Create_Call) Return(_a0 *corev1.ConfigMap, _a1 error) *ConfigMaps_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Delete provides a mock function with given fields: ctx, name, opts
func (_m *ConfigMaps) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	ret := _m.Called(ctx, name, opts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, metav1.DeleteOptions) error); ok {
		r0 = rf(ctx, name, opts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ConfigMaps_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type ConfigMaps_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
//   - opts metav1.DeleteOptions
func (_e *ConfigMaps_Expecter) Delete(ctx interface{}, name interface{}, opts interface{}) *ConfigMaps_Delete_Call {
	return &ConfigMaps_Delete_Call{Call: _e.mock.On("Delete", ctx, name, opts)}
}

func (_c *ConfigMaps_Delete_Call) Run(run func(ctx context.Context, name string, opts metav1.DeleteOptions)) *ConfigMaps_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(metav1.DeleteOptions))
	})
	return _c
}

func (_c *ConfigMaps_Delete_Call) Return(_a0 error) *ConfigMaps_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

// DeleteCollection provides a mock function with given fields: ctx, opts, listOpts
func (_m *ConfigMaps) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	ret := _m.Called(ctx, opts, listOpts)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, metav1.DeleteOptions, metav1.ListOptions) error); ok {
		r0 = rf(ctx, opts, listOpts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ConfigMaps_DeleteCollection_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteCollection'
type ConfigMaps_DeleteCollection_Call struct {
	*mock.Call
}

// DeleteCollection is a helper method to define mock.On call
//   - ctx context.Context
//   - opts metav1.DeleteOptions
//   - listOpts metav1.ListOptions
func (_e *ConfigMaps_Expecter) DeleteCollection(ctx interface{}, opts interface{}, listOpts interface{}) *ConfigMaps_DeleteCollection_Call {
	return &ConfigMaps_DeleteCollection_Call{Call: _e.mock.On("DeleteCollection", ctx, opts, listOpts)}
}

func (_c *ConfigMaps_DeleteCollection_Call) Run(run func(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions)) *ConfigMaps_DeleteCollection_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(metav1.DeleteOptions), args[2].(metav1.ListOptions))
	})
	return _c
}

func (_c *ConfigMaps_DeleteCollection_Call) Return(_a0 error) *ConfigMaps_DeleteCollection_Call {
	_c.Call.Return(_a0)
	return _c
}

// Get provides a mock function with given fields: ctx, name, opts
func (_m *ConfigMaps) Get(ctx context.Context, name string, opts metav1.GetOptions) (*corev1.ConfigMap, error) {
	ret := _m.Called(ctx, name, opts)

	var r0 *corev1.ConfigMap
	if rf, ok := ret.Get(0).(func(context.Context, string, metav1.GetOptions) *corev1.ConfigMap); ok {
		r0 = rf(ctx, name, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, metav1.GetOptions) error); ok {
		r1 = rf(ctx, name, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConfigMaps_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type ConfigMaps_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
//   - opts metav1.GetOptions
func (_e *ConfigMaps_Expecter) Get(ctx interface{}, name interface{}, opts interface{}) *ConfigMaps_Get_Call {
	return &ConfigMaps_Get_Call{Call: _e.mock.On("Get", ctx, name, opts)}
}

func (_c *ConfigMaps_Get_Call) Run(run func(ctx context.Context, name string, opts metav1.GetOptions)) *ConfigMaps_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(metav1.GetOptions))
	})
	return _c
}

func (_c *ConfigMaps_Get_Call) Return(_a0 *corev1.ConfigMap, _a1 error) *ConfigMaps_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// List provides a mock function with given fields: ctx, opts
func (_m *ConfigMaps) List(ctx context.Context, opts metav1.ListOptions) (*corev1.ConfigMapList, error) {
	ret := _m.Called(ctx, opts)

	var r0 *corev1.ConfigMapList
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) *corev1.ConfigMapList); ok {
		r0 = rf(ctx, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.ConfigMapList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, metav1.ListOptions) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConfigMaps_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type ConfigMaps_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
//   - opts metav1.ListOptions
func (_e *ConfigMaps_Expecter) List(ctx interface{}, opts interface{}) *ConfigMaps_List_Call {
	return &ConfigMaps_List_Call{Call: _e.mock.On("List", ctx, opts)}
}

func (_c *ConfigMaps_List_Call) Run(run func(ctx context.Context, opts metav1.ListOptions)) *ConfigMaps_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(metav1.ListOptions))
	})
	return _c
}

func (_c *ConfigMaps_List_Call) Return(_a0 *corev1.ConfigMapList, _a1 error) *ConfigMaps_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Patch provides a mock function with given fields: ctx, name, pt, data, opts, subresources
func (_m *ConfigMaps) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*corev1.ConfigMap, error) {
	_va := make([]interface{}, len(subresources))
	for _i := range subresources {
		_va[_i] = subresources[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, name, pt, data, opts)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *corev1.ConfigMap
	if rf, ok := ret.Get(0).(func(context.Context, string, types.PatchType, []byte, metav1.PatchOptions, ...string) *corev1.ConfigMap); ok {
		r0 = rf(ctx, name, pt, data, opts, subresources...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, types.PatchType, []byte, metav1.PatchOptions, ...string) error); ok {
		r1 = rf(ctx, name, pt, data, opts, subresources...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConfigMaps_Patch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Patch'
type ConfigMaps_Patch_Call struct {
	*mock.Call
}

// Patch is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
//   - pt types.PatchType
//   - data []byte
//   - opts metav1.PatchOptions
//   - subresources ...string
func (_e *ConfigMaps_Expecter) Patch(ctx interface{}, name interface{}, pt interface{}, data interface{}, opts interface{}, subresources ...interface{}) *ConfigMaps_Patch_Call {
	return &ConfigMaps_Patch_Call{Call: _e.mock.On("Patch",
		append([]interface{}{ctx, name, pt, data, opts}, subresources...)...)}
}

func (_c *ConfigMaps_Patch_Call) Run(run func(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string)) *ConfigMaps_Patch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]string, len(args)-5)
		for i, a := range args[5:] {
			if a != nil {
				variadicArgs[i] = a.(string)
			}
		}
		run(args[0].(context.Context), args[1].(string), args[2].(types.PatchType), args[3].([]byte), args[4].(metav1.PatchOptions), variadicArgs...)
	})
	return _c
}

func (_c *ConfigMaps_Patch_Call) Return(result *corev1.ConfigMap, err error) *ConfigMaps_Patch_Call {
	_c.Call.Return(result, err)
	return _c
}

// Update provides a mock function with given fields: ctx, configMap, opts
func (_m *ConfigMaps) Update(ctx context.Context, configMap *corev1.ConfigMap, opts metav1.UpdateOptions) (*corev1.ConfigMap, error) {
	ret := _m.Called(ctx, configMap, opts)

	var r0 *corev1.ConfigMap
	if rf, ok := ret.Get(0).(func(context.Context, *corev1.ConfigMap, metav1.UpdateOptions) *corev1.ConfigMap); ok {
		r0 = rf(ctx, configMap, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *corev1.ConfigMap, metav1.UpdateOptions) error); ok {
		r1 = rf(ctx, configMap, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConfigMaps_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type ConfigMaps_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx context.Context
//   - configMap *corev1.ConfigMap
//   - opts metav1.UpdateOptions
func (_e *ConfigMaps_Expecter) Update(ctx interface{}, configMap interface{}, opts interface{}) *ConfigMaps_Update_Call {
	return &ConfigMaps_Update_Call{Call: _e.mock.On("Update", ctx, configMap, opts)}
}

func (_c *ConfigMaps_Update_Call) Run(run func(ctx context.Context, configMap *corev1.ConfigMap, opts metav1.UpdateOptions)) *ConfigMaps_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*corev1.ConfigMap), args[2].(metav1.UpdateOptions))
	})
	return _c
}

func (_c *ConfigMaps_Update_Call) Return(_a0 *corev1.ConfigMap, _a1 error) *ConfigMaps_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

// Watch provides a mock function with given fields: ctx, opts
func (_m *ConfigMaps) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	ret := _m.Called(ctx, opts)

	var r0 watch.Interface
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) watch.Interface); ok {
		r0 = rf(ctx, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(watch.Interface)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, metav1.ListOptions) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ConfigMaps_Watch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Watch'
type ConfigMaps_Watch_Call struct {
	*mock.Call
}

// Watch is a helper method to define mock.On call
//   - ctx context.Context
//   - opts metav1.ListOptions
func (_e *ConfigMaps_Expecter) Watch(ctx interface{}, opts interface{}) *ConfigMaps_Watch_Call {
	return &ConfigMaps_Watch_Call{Call: _e.mock.On("Watch", ctx, opts)}
}

func (_c *ConfigMaps_Watch_Call) Run(run func(ctx context.Context, opts metav1.ListOptions)) *ConfigMaps_Watch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(metav1.ListOptions))
	})
	return _c
}

func (_c *ConfigMaps_Watch_Call) Return(_a0 watch.Interface, _a1 error) *ConfigMaps_Watch_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

type mockConstructorTestingTNewConfigMaps interface {
	mock.TestingT
	Cleanup(func())
}

// NewConfigMaps creates a new instance of ConfigMaps. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewConfigMaps(t mockConstructorTestingTNewConfigMaps) *ConfigMaps {
	mock := &ConfigMaps{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}