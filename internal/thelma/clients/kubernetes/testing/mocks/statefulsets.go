// Code generated by mockery v2.40.1. DO NOT EDIT.

package mocks

import (
	appsv1 "k8s.io/api/apps/v1"
	apiautoscalingv1 "k8s.io/api/autoscaling/v1"

	autoscalingv1 "k8s.io/client-go/applyconfigurations/autoscaling/v1"

	context "context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mock "github.com/stretchr/testify/mock"

	types "k8s.io/apimachinery/pkg/types"

	v1 "k8s.io/client-go/applyconfigurations/apps/v1"

	watch "k8s.io/apimachinery/pkg/watch"
)

// StatefulSets is an autogenerated mock type for the StatefulSets type
type StatefulSets struct {
	mock.Mock
}

type StatefulSets_Expecter struct {
	mock *mock.Mock
}

func (_m *StatefulSets) EXPECT() *StatefulSets_Expecter {
	return &StatefulSets_Expecter{mock: &_m.Mock}
}

// Apply provides a mock function with given fields: ctx, statefulSet, opts
func (_m *StatefulSets) Apply(ctx context.Context, statefulSet *v1.StatefulSetApplyConfiguration, opts metav1.ApplyOptions) (*appsv1.StatefulSet, error) {
	ret := _m.Called(ctx, statefulSet, opts)

	if len(ret) == 0 {
		panic("no return value specified for Apply")
	}

	var r0 *appsv1.StatefulSet
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.StatefulSetApplyConfiguration, metav1.ApplyOptions) (*appsv1.StatefulSet, error)); ok {
		return rf(ctx, statefulSet, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.StatefulSetApplyConfiguration, metav1.ApplyOptions) *appsv1.StatefulSet); ok {
		r0 = rf(ctx, statefulSet, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*appsv1.StatefulSet)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.StatefulSetApplyConfiguration, metav1.ApplyOptions) error); ok {
		r1 = rf(ctx, statefulSet, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StatefulSets_Apply_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Apply'
type StatefulSets_Apply_Call struct {
	*mock.Call
}

// Apply is a helper method to define mock.On call
//   - ctx context.Context
//   - statefulSet *v1.StatefulSetApplyConfiguration
//   - opts metav1.ApplyOptions
func (_e *StatefulSets_Expecter) Apply(ctx interface{}, statefulSet interface{}, opts interface{}) *StatefulSets_Apply_Call {
	return &StatefulSets_Apply_Call{Call: _e.mock.On("Apply", ctx, statefulSet, opts)}
}

func (_c *StatefulSets_Apply_Call) Run(run func(ctx context.Context, statefulSet *v1.StatefulSetApplyConfiguration, opts metav1.ApplyOptions)) *StatefulSets_Apply_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v1.StatefulSetApplyConfiguration), args[2].(metav1.ApplyOptions))
	})
	return _c
}

func (_c *StatefulSets_Apply_Call) Return(result *appsv1.StatefulSet, err error) *StatefulSets_Apply_Call {
	_c.Call.Return(result, err)
	return _c
}

func (_c *StatefulSets_Apply_Call) RunAndReturn(run func(context.Context, *v1.StatefulSetApplyConfiguration, metav1.ApplyOptions) (*appsv1.StatefulSet, error)) *StatefulSets_Apply_Call {
	_c.Call.Return(run)
	return _c
}

// ApplyScale provides a mock function with given fields: ctx, statefulSetName, scale, opts
func (_m *StatefulSets) ApplyScale(ctx context.Context, statefulSetName string, scale *autoscalingv1.ScaleApplyConfiguration, opts metav1.ApplyOptions) (*apiautoscalingv1.Scale, error) {
	ret := _m.Called(ctx, statefulSetName, scale, opts)

	if len(ret) == 0 {
		panic("no return value specified for ApplyScale")
	}

	var r0 *apiautoscalingv1.Scale
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *autoscalingv1.ScaleApplyConfiguration, metav1.ApplyOptions) (*apiautoscalingv1.Scale, error)); ok {
		return rf(ctx, statefulSetName, scale, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *autoscalingv1.ScaleApplyConfiguration, metav1.ApplyOptions) *apiautoscalingv1.Scale); ok {
		r0 = rf(ctx, statefulSetName, scale, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*apiautoscalingv1.Scale)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *autoscalingv1.ScaleApplyConfiguration, metav1.ApplyOptions) error); ok {
		r1 = rf(ctx, statefulSetName, scale, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StatefulSets_ApplyScale_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ApplyScale'
type StatefulSets_ApplyScale_Call struct {
	*mock.Call
}

// ApplyScale is a helper method to define mock.On call
//   - ctx context.Context
//   - statefulSetName string
//   - scale *autoscalingv1.ScaleApplyConfiguration
//   - opts metav1.ApplyOptions
func (_e *StatefulSets_Expecter) ApplyScale(ctx interface{}, statefulSetName interface{}, scale interface{}, opts interface{}) *StatefulSets_ApplyScale_Call {
	return &StatefulSets_ApplyScale_Call{Call: _e.mock.On("ApplyScale", ctx, statefulSetName, scale, opts)}
}

func (_c *StatefulSets_ApplyScale_Call) Run(run func(ctx context.Context, statefulSetName string, scale *autoscalingv1.ScaleApplyConfiguration, opts metav1.ApplyOptions)) *StatefulSets_ApplyScale_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*autoscalingv1.ScaleApplyConfiguration), args[3].(metav1.ApplyOptions))
	})
	return _c
}

func (_c *StatefulSets_ApplyScale_Call) Return(_a0 *apiautoscalingv1.Scale, _a1 error) *StatefulSets_ApplyScale_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StatefulSets_ApplyScale_Call) RunAndReturn(run func(context.Context, string, *autoscalingv1.ScaleApplyConfiguration, metav1.ApplyOptions) (*apiautoscalingv1.Scale, error)) *StatefulSets_ApplyScale_Call {
	_c.Call.Return(run)
	return _c
}

// ApplyStatus provides a mock function with given fields: ctx, statefulSet, opts
func (_m *StatefulSets) ApplyStatus(ctx context.Context, statefulSet *v1.StatefulSetApplyConfiguration, opts metav1.ApplyOptions) (*appsv1.StatefulSet, error) {
	ret := _m.Called(ctx, statefulSet, opts)

	if len(ret) == 0 {
		panic("no return value specified for ApplyStatus")
	}

	var r0 *appsv1.StatefulSet
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *v1.StatefulSetApplyConfiguration, metav1.ApplyOptions) (*appsv1.StatefulSet, error)); ok {
		return rf(ctx, statefulSet, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *v1.StatefulSetApplyConfiguration, metav1.ApplyOptions) *appsv1.StatefulSet); ok {
		r0 = rf(ctx, statefulSet, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*appsv1.StatefulSet)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *v1.StatefulSetApplyConfiguration, metav1.ApplyOptions) error); ok {
		r1 = rf(ctx, statefulSet, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StatefulSets_ApplyStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ApplyStatus'
type StatefulSets_ApplyStatus_Call struct {
	*mock.Call
}

// ApplyStatus is a helper method to define mock.On call
//   - ctx context.Context
//   - statefulSet *v1.StatefulSetApplyConfiguration
//   - opts metav1.ApplyOptions
func (_e *StatefulSets_Expecter) ApplyStatus(ctx interface{}, statefulSet interface{}, opts interface{}) *StatefulSets_ApplyStatus_Call {
	return &StatefulSets_ApplyStatus_Call{Call: _e.mock.On("ApplyStatus", ctx, statefulSet, opts)}
}

func (_c *StatefulSets_ApplyStatus_Call) Run(run func(ctx context.Context, statefulSet *v1.StatefulSetApplyConfiguration, opts metav1.ApplyOptions)) *StatefulSets_ApplyStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*v1.StatefulSetApplyConfiguration), args[2].(metav1.ApplyOptions))
	})
	return _c
}

func (_c *StatefulSets_ApplyStatus_Call) Return(result *appsv1.StatefulSet, err error) *StatefulSets_ApplyStatus_Call {
	_c.Call.Return(result, err)
	return _c
}

func (_c *StatefulSets_ApplyStatus_Call) RunAndReturn(run func(context.Context, *v1.StatefulSetApplyConfiguration, metav1.ApplyOptions) (*appsv1.StatefulSet, error)) *StatefulSets_ApplyStatus_Call {
	_c.Call.Return(run)
	return _c
}

// Create provides a mock function with given fields: ctx, statefulSet, opts
func (_m *StatefulSets) Create(ctx context.Context, statefulSet *appsv1.StatefulSet, opts metav1.CreateOptions) (*appsv1.StatefulSet, error) {
	ret := _m.Called(ctx, statefulSet, opts)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

	var r0 *appsv1.StatefulSet
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *appsv1.StatefulSet, metav1.CreateOptions) (*appsv1.StatefulSet, error)); ok {
		return rf(ctx, statefulSet, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *appsv1.StatefulSet, metav1.CreateOptions) *appsv1.StatefulSet); ok {
		r0 = rf(ctx, statefulSet, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*appsv1.StatefulSet)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *appsv1.StatefulSet, metav1.CreateOptions) error); ok {
		r1 = rf(ctx, statefulSet, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StatefulSets_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type StatefulSets_Create_Call struct {
	*mock.Call
}

// Create is a helper method to define mock.On call
//   - ctx context.Context
//   - statefulSet *appsv1.StatefulSet
//   - opts metav1.CreateOptions
func (_e *StatefulSets_Expecter) Create(ctx interface{}, statefulSet interface{}, opts interface{}) *StatefulSets_Create_Call {
	return &StatefulSets_Create_Call{Call: _e.mock.On("Create", ctx, statefulSet, opts)}
}

func (_c *StatefulSets_Create_Call) Run(run func(ctx context.Context, statefulSet *appsv1.StatefulSet, opts metav1.CreateOptions)) *StatefulSets_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*appsv1.StatefulSet), args[2].(metav1.CreateOptions))
	})
	return _c
}

func (_c *StatefulSets_Create_Call) Return(_a0 *appsv1.StatefulSet, _a1 error) *StatefulSets_Create_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StatefulSets_Create_Call) RunAndReturn(run func(context.Context, *appsv1.StatefulSet, metav1.CreateOptions) (*appsv1.StatefulSet, error)) *StatefulSets_Create_Call {
	_c.Call.Return(run)
	return _c
}

// Delete provides a mock function with given fields: ctx, name, opts
func (_m *StatefulSets) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	ret := _m.Called(ctx, name, opts)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, metav1.DeleteOptions) error); ok {
		r0 = rf(ctx, name, opts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StatefulSets_Delete_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Delete'
type StatefulSets_Delete_Call struct {
	*mock.Call
}

// Delete is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
//   - opts metav1.DeleteOptions
func (_e *StatefulSets_Expecter) Delete(ctx interface{}, name interface{}, opts interface{}) *StatefulSets_Delete_Call {
	return &StatefulSets_Delete_Call{Call: _e.mock.On("Delete", ctx, name, opts)}
}

func (_c *StatefulSets_Delete_Call) Run(run func(ctx context.Context, name string, opts metav1.DeleteOptions)) *StatefulSets_Delete_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(metav1.DeleteOptions))
	})
	return _c
}

func (_c *StatefulSets_Delete_Call) Return(_a0 error) *StatefulSets_Delete_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *StatefulSets_Delete_Call) RunAndReturn(run func(context.Context, string, metav1.DeleteOptions) error) *StatefulSets_Delete_Call {
	_c.Call.Return(run)
	return _c
}

// DeleteCollection provides a mock function with given fields: ctx, opts, listOpts
func (_m *StatefulSets) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	ret := _m.Called(ctx, opts, listOpts)

	if len(ret) == 0 {
		panic("no return value specified for DeleteCollection")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, metav1.DeleteOptions, metav1.ListOptions) error); ok {
		r0 = rf(ctx, opts, listOpts)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StatefulSets_DeleteCollection_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteCollection'
type StatefulSets_DeleteCollection_Call struct {
	*mock.Call
}

// DeleteCollection is a helper method to define mock.On call
//   - ctx context.Context
//   - opts metav1.DeleteOptions
//   - listOpts metav1.ListOptions
func (_e *StatefulSets_Expecter) DeleteCollection(ctx interface{}, opts interface{}, listOpts interface{}) *StatefulSets_DeleteCollection_Call {
	return &StatefulSets_DeleteCollection_Call{Call: _e.mock.On("DeleteCollection", ctx, opts, listOpts)}
}

func (_c *StatefulSets_DeleteCollection_Call) Run(run func(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions)) *StatefulSets_DeleteCollection_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(metav1.DeleteOptions), args[2].(metav1.ListOptions))
	})
	return _c
}

func (_c *StatefulSets_DeleteCollection_Call) Return(_a0 error) *StatefulSets_DeleteCollection_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *StatefulSets_DeleteCollection_Call) RunAndReturn(run func(context.Context, metav1.DeleteOptions, metav1.ListOptions) error) *StatefulSets_DeleteCollection_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, name, opts
func (_m *StatefulSets) Get(ctx context.Context, name string, opts metav1.GetOptions) (*appsv1.StatefulSet, error) {
	ret := _m.Called(ctx, name, opts)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 *appsv1.StatefulSet
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, metav1.GetOptions) (*appsv1.StatefulSet, error)); ok {
		return rf(ctx, name, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, metav1.GetOptions) *appsv1.StatefulSet); ok {
		r0 = rf(ctx, name, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*appsv1.StatefulSet)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, metav1.GetOptions) error); ok {
		r1 = rf(ctx, name, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StatefulSets_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type StatefulSets_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
//   - opts metav1.GetOptions
func (_e *StatefulSets_Expecter) Get(ctx interface{}, name interface{}, opts interface{}) *StatefulSets_Get_Call {
	return &StatefulSets_Get_Call{Call: _e.mock.On("Get", ctx, name, opts)}
}

func (_c *StatefulSets_Get_Call) Run(run func(ctx context.Context, name string, opts metav1.GetOptions)) *StatefulSets_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(metav1.GetOptions))
	})
	return _c
}

func (_c *StatefulSets_Get_Call) Return(_a0 *appsv1.StatefulSet, _a1 error) *StatefulSets_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StatefulSets_Get_Call) RunAndReturn(run func(context.Context, string, metav1.GetOptions) (*appsv1.StatefulSet, error)) *StatefulSets_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetScale provides a mock function with given fields: ctx, statefulSetName, options
func (_m *StatefulSets) GetScale(ctx context.Context, statefulSetName string, options metav1.GetOptions) (*apiautoscalingv1.Scale, error) {
	ret := _m.Called(ctx, statefulSetName, options)

	if len(ret) == 0 {
		panic("no return value specified for GetScale")
	}

	var r0 *apiautoscalingv1.Scale
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, metav1.GetOptions) (*apiautoscalingv1.Scale, error)); ok {
		return rf(ctx, statefulSetName, options)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, metav1.GetOptions) *apiautoscalingv1.Scale); ok {
		r0 = rf(ctx, statefulSetName, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*apiautoscalingv1.Scale)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, metav1.GetOptions) error); ok {
		r1 = rf(ctx, statefulSetName, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StatefulSets_GetScale_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetScale'
type StatefulSets_GetScale_Call struct {
	*mock.Call
}

// GetScale is a helper method to define mock.On call
//   - ctx context.Context
//   - statefulSetName string
//   - options metav1.GetOptions
func (_e *StatefulSets_Expecter) GetScale(ctx interface{}, statefulSetName interface{}, options interface{}) *StatefulSets_GetScale_Call {
	return &StatefulSets_GetScale_Call{Call: _e.mock.On("GetScale", ctx, statefulSetName, options)}
}

func (_c *StatefulSets_GetScale_Call) Run(run func(ctx context.Context, statefulSetName string, options metav1.GetOptions)) *StatefulSets_GetScale_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(metav1.GetOptions))
	})
	return _c
}

func (_c *StatefulSets_GetScale_Call) Return(_a0 *apiautoscalingv1.Scale, _a1 error) *StatefulSets_GetScale_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StatefulSets_GetScale_Call) RunAndReturn(run func(context.Context, string, metav1.GetOptions) (*apiautoscalingv1.Scale, error)) *StatefulSets_GetScale_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with given fields: ctx, opts
func (_m *StatefulSets) List(ctx context.Context, opts metav1.ListOptions) (*appsv1.StatefulSetList, error) {
	ret := _m.Called(ctx, opts)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 *appsv1.StatefulSetList
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) (*appsv1.StatefulSetList, error)); ok {
		return rf(ctx, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) *appsv1.StatefulSetList); ok {
		r0 = rf(ctx, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*appsv1.StatefulSetList)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, metav1.ListOptions) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StatefulSets_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type StatefulSets_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
//   - opts metav1.ListOptions
func (_e *StatefulSets_Expecter) List(ctx interface{}, opts interface{}) *StatefulSets_List_Call {
	return &StatefulSets_List_Call{Call: _e.mock.On("List", ctx, opts)}
}

func (_c *StatefulSets_List_Call) Run(run func(ctx context.Context, opts metav1.ListOptions)) *StatefulSets_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(metav1.ListOptions))
	})
	return _c
}

func (_c *StatefulSets_List_Call) Return(_a0 *appsv1.StatefulSetList, _a1 error) *StatefulSets_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StatefulSets_List_Call) RunAndReturn(run func(context.Context, metav1.ListOptions) (*appsv1.StatefulSetList, error)) *StatefulSets_List_Call {
	_c.Call.Return(run)
	return _c
}

// Patch provides a mock function with given fields: ctx, name, pt, data, opts, subresources
func (_m *StatefulSets) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (*appsv1.StatefulSet, error) {
	_va := make([]interface{}, len(subresources))
	for _i := range subresources {
		_va[_i] = subresources[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, name, pt, data, opts)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Patch")
	}

	var r0 *appsv1.StatefulSet
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, types.PatchType, []byte, metav1.PatchOptions, ...string) (*appsv1.StatefulSet, error)); ok {
		return rf(ctx, name, pt, data, opts, subresources...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, types.PatchType, []byte, metav1.PatchOptions, ...string) *appsv1.StatefulSet); ok {
		r0 = rf(ctx, name, pt, data, opts, subresources...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*appsv1.StatefulSet)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, types.PatchType, []byte, metav1.PatchOptions, ...string) error); ok {
		r1 = rf(ctx, name, pt, data, opts, subresources...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StatefulSets_Patch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Patch'
type StatefulSets_Patch_Call struct {
	*mock.Call
}

// Patch is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
//   - pt types.PatchType
//   - data []byte
//   - opts metav1.PatchOptions
//   - subresources ...string
func (_e *StatefulSets_Expecter) Patch(ctx interface{}, name interface{}, pt interface{}, data interface{}, opts interface{}, subresources ...interface{}) *StatefulSets_Patch_Call {
	return &StatefulSets_Patch_Call{Call: _e.mock.On("Patch",
		append([]interface{}{ctx, name, pt, data, opts}, subresources...)...)}
}

func (_c *StatefulSets_Patch_Call) Run(run func(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string)) *StatefulSets_Patch_Call {
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

func (_c *StatefulSets_Patch_Call) Return(result *appsv1.StatefulSet, err error) *StatefulSets_Patch_Call {
	_c.Call.Return(result, err)
	return _c
}

func (_c *StatefulSets_Patch_Call) RunAndReturn(run func(context.Context, string, types.PatchType, []byte, metav1.PatchOptions, ...string) (*appsv1.StatefulSet, error)) *StatefulSets_Patch_Call {
	_c.Call.Return(run)
	return _c
}

// Update provides a mock function with given fields: ctx, statefulSet, opts
func (_m *StatefulSets) Update(ctx context.Context, statefulSet *appsv1.StatefulSet, opts metav1.UpdateOptions) (*appsv1.StatefulSet, error) {
	ret := _m.Called(ctx, statefulSet, opts)

	if len(ret) == 0 {
		panic("no return value specified for Update")
	}

	var r0 *appsv1.StatefulSet
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *appsv1.StatefulSet, metav1.UpdateOptions) (*appsv1.StatefulSet, error)); ok {
		return rf(ctx, statefulSet, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *appsv1.StatefulSet, metav1.UpdateOptions) *appsv1.StatefulSet); ok {
		r0 = rf(ctx, statefulSet, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*appsv1.StatefulSet)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *appsv1.StatefulSet, metav1.UpdateOptions) error); ok {
		r1 = rf(ctx, statefulSet, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StatefulSets_Update_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Update'
type StatefulSets_Update_Call struct {
	*mock.Call
}

// Update is a helper method to define mock.On call
//   - ctx context.Context
//   - statefulSet *appsv1.StatefulSet
//   - opts metav1.UpdateOptions
func (_e *StatefulSets_Expecter) Update(ctx interface{}, statefulSet interface{}, opts interface{}) *StatefulSets_Update_Call {
	return &StatefulSets_Update_Call{Call: _e.mock.On("Update", ctx, statefulSet, opts)}
}

func (_c *StatefulSets_Update_Call) Run(run func(ctx context.Context, statefulSet *appsv1.StatefulSet, opts metav1.UpdateOptions)) *StatefulSets_Update_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*appsv1.StatefulSet), args[2].(metav1.UpdateOptions))
	})
	return _c
}

func (_c *StatefulSets_Update_Call) Return(_a0 *appsv1.StatefulSet, _a1 error) *StatefulSets_Update_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StatefulSets_Update_Call) RunAndReturn(run func(context.Context, *appsv1.StatefulSet, metav1.UpdateOptions) (*appsv1.StatefulSet, error)) *StatefulSets_Update_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateScale provides a mock function with given fields: ctx, statefulSetName, scale, opts
func (_m *StatefulSets) UpdateScale(ctx context.Context, statefulSetName string, scale *apiautoscalingv1.Scale, opts metav1.UpdateOptions) (*apiautoscalingv1.Scale, error) {
	ret := _m.Called(ctx, statefulSetName, scale, opts)

	if len(ret) == 0 {
		panic("no return value specified for UpdateScale")
	}

	var r0 *apiautoscalingv1.Scale
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *apiautoscalingv1.Scale, metav1.UpdateOptions) (*apiautoscalingv1.Scale, error)); ok {
		return rf(ctx, statefulSetName, scale, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *apiautoscalingv1.Scale, metav1.UpdateOptions) *apiautoscalingv1.Scale); ok {
		r0 = rf(ctx, statefulSetName, scale, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*apiautoscalingv1.Scale)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *apiautoscalingv1.Scale, metav1.UpdateOptions) error); ok {
		r1 = rf(ctx, statefulSetName, scale, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StatefulSets_UpdateScale_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateScale'
type StatefulSets_UpdateScale_Call struct {
	*mock.Call
}

// UpdateScale is a helper method to define mock.On call
//   - ctx context.Context
//   - statefulSetName string
//   - scale *apiautoscalingv1.Scale
//   - opts metav1.UpdateOptions
func (_e *StatefulSets_Expecter) UpdateScale(ctx interface{}, statefulSetName interface{}, scale interface{}, opts interface{}) *StatefulSets_UpdateScale_Call {
	return &StatefulSets_UpdateScale_Call{Call: _e.mock.On("UpdateScale", ctx, statefulSetName, scale, opts)}
}

func (_c *StatefulSets_UpdateScale_Call) Run(run func(ctx context.Context, statefulSetName string, scale *apiautoscalingv1.Scale, opts metav1.UpdateOptions)) *StatefulSets_UpdateScale_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(*apiautoscalingv1.Scale), args[3].(metav1.UpdateOptions))
	})
	return _c
}

func (_c *StatefulSets_UpdateScale_Call) Return(_a0 *apiautoscalingv1.Scale, _a1 error) *StatefulSets_UpdateScale_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StatefulSets_UpdateScale_Call) RunAndReturn(run func(context.Context, string, *apiautoscalingv1.Scale, metav1.UpdateOptions) (*apiautoscalingv1.Scale, error)) *StatefulSets_UpdateScale_Call {
	_c.Call.Return(run)
	return _c
}

// UpdateStatus provides a mock function with given fields: ctx, statefulSet, opts
func (_m *StatefulSets) UpdateStatus(ctx context.Context, statefulSet *appsv1.StatefulSet, opts metav1.UpdateOptions) (*appsv1.StatefulSet, error) {
	ret := _m.Called(ctx, statefulSet, opts)

	if len(ret) == 0 {
		panic("no return value specified for UpdateStatus")
	}

	var r0 *appsv1.StatefulSet
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *appsv1.StatefulSet, metav1.UpdateOptions) (*appsv1.StatefulSet, error)); ok {
		return rf(ctx, statefulSet, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *appsv1.StatefulSet, metav1.UpdateOptions) *appsv1.StatefulSet); ok {
		r0 = rf(ctx, statefulSet, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*appsv1.StatefulSet)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *appsv1.StatefulSet, metav1.UpdateOptions) error); ok {
		r1 = rf(ctx, statefulSet, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StatefulSets_UpdateStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UpdateStatus'
type StatefulSets_UpdateStatus_Call struct {
	*mock.Call
}

// UpdateStatus is a helper method to define mock.On call
//   - ctx context.Context
//   - statefulSet *appsv1.StatefulSet
//   - opts metav1.UpdateOptions
func (_e *StatefulSets_Expecter) UpdateStatus(ctx interface{}, statefulSet interface{}, opts interface{}) *StatefulSets_UpdateStatus_Call {
	return &StatefulSets_UpdateStatus_Call{Call: _e.mock.On("UpdateStatus", ctx, statefulSet, opts)}
}

func (_c *StatefulSets_UpdateStatus_Call) Run(run func(ctx context.Context, statefulSet *appsv1.StatefulSet, opts metav1.UpdateOptions)) *StatefulSets_UpdateStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*appsv1.StatefulSet), args[2].(metav1.UpdateOptions))
	})
	return _c
}

func (_c *StatefulSets_UpdateStatus_Call) Return(_a0 *appsv1.StatefulSet, _a1 error) *StatefulSets_UpdateStatus_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StatefulSets_UpdateStatus_Call) RunAndReturn(run func(context.Context, *appsv1.StatefulSet, metav1.UpdateOptions) (*appsv1.StatefulSet, error)) *StatefulSets_UpdateStatus_Call {
	_c.Call.Return(run)
	return _c
}

// Watch provides a mock function with given fields: ctx, opts
func (_m *StatefulSets) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	ret := _m.Called(ctx, opts)

	if len(ret) == 0 {
		panic("no return value specified for Watch")
	}

	var r0 watch.Interface
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) (watch.Interface, error)); ok {
		return rf(ctx, opts)
	}
	if rf, ok := ret.Get(0).(func(context.Context, metav1.ListOptions) watch.Interface); ok {
		r0 = rf(ctx, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(watch.Interface)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, metav1.ListOptions) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// StatefulSets_Watch_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Watch'
type StatefulSets_Watch_Call struct {
	*mock.Call
}

// Watch is a helper method to define mock.On call
//   - ctx context.Context
//   - opts metav1.ListOptions
func (_e *StatefulSets_Expecter) Watch(ctx interface{}, opts interface{}) *StatefulSets_Watch_Call {
	return &StatefulSets_Watch_Call{Call: _e.mock.On("Watch", ctx, opts)}
}

func (_c *StatefulSets_Watch_Call) Run(run func(ctx context.Context, opts metav1.ListOptions)) *StatefulSets_Watch_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(metav1.ListOptions))
	})
	return _c
}

func (_c *StatefulSets_Watch_Call) Return(_a0 watch.Interface, _a1 error) *StatefulSets_Watch_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *StatefulSets_Watch_Call) RunAndReturn(run func(context.Context, metav1.ListOptions) (watch.Interface, error)) *StatefulSets_Watch_Call {
	_c.Call.Return(run)
	return _c
}

// NewStatefulSets creates a new instance of StatefulSets. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStatefulSets(t interface {
	mock.TestingT
	Cleanup(func())
}) *StatefulSets {
	mock := &StatefulSets{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
