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
	"alex-opr/controllers/tools"
	"context"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"log"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
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
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch

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

	/*
		获取调协的crd实例
	*/
	// 1 获取集群中的资源对象
	// 实例出一个空的对象
	ins := &mockv1beta1.MacBook{}
	// client/Reader 接口 调用get方法从api中获取创建的对象
	err := r.Get(context.TODO(), req.NamespacedName, ins)
	if err != nil {
		return ctrl.Result{}, err
	} else {
		r.Recorder.Event(ins, "Normal", "GetRes", "获取自定义资源成功")
	}

	/*
		finalizers 处理
	*/
	//
	// name of our custom finalizer
	myFinalizerName := "dong.com/finalizer"

	// examine DeletionTimestamp to determine if object is under deletion
	if ins.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(ins.GetFinalizers(), myFinalizerName) {
			ins.SetFinalizers(append(ins.GetFinalizers(), myFinalizerName))
			if err := r.Update(ctx, ins); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(ins.GetFinalizers(), myFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(ins); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			ins.SetFinalizers(removeString(ins.GetFinalizers(), myFinalizerName))
			if err := r.Update(ctx, ins); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	/*
		创建dep并建立关系
	*/
	// new一个dep对象,是个指针类型 *dep
	//dep := tools.NewCreatedep(ins)

	dep := tools.NewDeployMent(ins)

	/*
		建立关系
	*/
	err = controllerutil.SetOwnerReference(ins, dep, r.Scheme)
	if err != nil {
		log.Println("set fiald")
	}

	// 将查找的对象填入下面的指针类型变量中
	found := &appsv1.Deployment{}
	// 在集群中查找dep对象
	// type ObjectKey types.NamespacedName
	// Object 需要是一个指针类型
	// 找到了就为空
	err = r.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, found)

	if err != nil {
		// 2 调用 client/Writer 接口来往k8s里面创建资源
		err = r.Create(context.TODO(), dep)
		if err != nil {
			log.Println("dep created not ok")
		} else {
			log.Println("dep 创建成功")
		}
	}

	/*
			再添加了 own 方法后


		depp := &appsv1.Deployment{}
		err = r.Get(context.TODO(), req.NamespacedName, depp)
		if err != nil {
			fmt.Println("depp 么有发现")

			//return ctrl.Result{}, err
		}

		fmt.Println("depp 发现了")

		ins.Labels["pod-count"] = fmt.Sprintf("%v", 8)
		err = r.Update(ctx, ins)
		if err != nil {
			return reconcile.Result{}, err
		}
	*/

	err = r.Update(ctx, ins)
	if err != nil {
		log.Println("更新失败")
	}

	return ctrl.Result{}, nil

}

func (r *MacBookReconciler) InjectClient(c client.Client) error {
	r.Client = c
	return nil
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func (r *MacBookReconciler) deleteExternalResources(macbook *mockv1beta1.MacBook) error {
	//
	// delete any external resources associated with the ins
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple times for same object.
	time.Sleep(1 * time.Second)
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MacBookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// for指定需要监听的资源 基于watch实现
		// Watches(&source.Kind{Type: apiType}, &handler.EnqueueRequestForObject{})
		For(&mockv1beta1.MacBook{}).
		// 指定监听crd的子资源
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
