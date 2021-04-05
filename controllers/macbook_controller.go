/*
Copyright 2021 lirui.

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

package controllers

import (
	mockv1beta1 "alex-opr/api/v1beta1"
	"alex-opr/controllers/create_res"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// MacBookReconciler 是个框架可以按需添加相关功能
// MacBookReconciler reconciles a MacBook object
type MacBookReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	// 添加事件记录器
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=mock.dong.com,resources=macbooks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mock.dong.com,resources=macbooks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mock.dong.com,resources=macbooks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MacBook object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *MacBookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("macbook", req.NamespacedName)

	// 1 获取集群中的资源对象
	// 实例出一个空的对象
	ins := &mockv1beta1.MacBook{}
	// client/Reader 接口 调用get方法从api中获取创建的对象
	err := r.Get(context.TODO(), req.NamespacedName, ins)
	if err != nil {
		return ctrl.Result{}, err
	} else {
		fmt.Printf("MacBook res get [%v] \n", ins.Spec.DisPlay)
		r.Recorder.Event(ins, "Normal", "GetRes", "okok")
	}

	// new一个pod对象,是个指针类型 *pod
	pod := create_res.NewCreatePod(ins)

	// 建立关联关系
	err = controllerutil.SetOwnerReference(ins, pod, r.Scheme)

	if err != nil {
		fmt.Print("set not ok\n")
	}

	// 将查找的对象填入下面的指针类型变量中
	found := &corev1.Pod{}
	// 在集群中查找pod对象
	// type ObjectKey types.NamespacedName
	// Object 需要是一个指针类型
	// 找到了就为空
	err = r.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)

	if err != nil {
		// 2 调用 client/Writer 接口来往k8s里面创建资源
		err = r.Create(context.TODO(), pod)
		if err != nil {
			fmt.Printf("pod %v create fail,err is %v \n", pod.Name, err)
		} else {
			fmt.Printf("pod %v create ok \n", pod.Name)
		}
	} else {
		fmt.Printf("pod [%v] has create\n", pod.Name)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MacBookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mockv1beta1.MacBook{}).
		Complete(r)
}
