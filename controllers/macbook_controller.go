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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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

// 注意权限管理，进行相关权限给予

//+kubebuilder:rbac:groups=mock.dong.com,resources=macbooks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mock.dong.com,resources=macbooks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mock.dong.com,resources=macbooks/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch

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
	clog := r.Log.WithValues("macbook", req.NamespacedName)

	/*
		获取调协的crd实例
	*/
	// 1 获取集群中的资源对象
	// 实例出一个空的对象
	MacBook := &mockv1beta1.MacBook{}
	// client/Reader 接口 调用get方法从api中获取创建的对象
	err := r.Get(context.TODO(), req.NamespacedName, MacBook)
	if err != nil {
		return ctrl.Result{}, err
	} else {
		clog.Info("find MacBook !", "MacBook-Annotations", MacBook.Annotations)
	}

	r.Recorder.Event(MacBook, "Normal", "BeginReconcile", "开始调协了")

	/*
		finalizers 处理
		示例代码 https://github.com/kubernetes-sigs/kubebuilder/blob/0317c63acfc2fb55a61492817968f09c4f7e20fa/docs/book/src/cronjob-tutorial/testdata/finalizer_example.go#L54
	*/
	//
	// name of our custom finalizer
	myFinalizerName := "dong.com/finalizer"

	// examine DeletionTimestamp to determine if object is under deletion
	if MacBook.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(MacBook.GetFinalizers(), myFinalizerName) {
			MacBook.SetFinalizers(append(MacBook.GetFinalizers(), myFinalizerName))
			if err := r.Update(ctx, MacBook); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(MacBook.GetFinalizers(), myFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			if err := r.deleteExternalResources(MacBook); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}

			// remove our finalizer from the list and update it.
			MacBook.SetFinalizers(removeString(MacBook.GetFinalizers(), myFinalizerName))
			if err := r.Update(ctx, MacBook); err != nil {
				return ctrl.Result{}, err
			}
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	/*
		创建dep并建立关系
	*/

	dep := tools.NewDeployMent(MacBook)

	/*
		建立关系
	*/
	err = controllerutil.SetControllerReference(MacBook, dep, r.Scheme)
	if err != nil {
		clog.Error(err, "SetControllerReference fail")
	}

	// 将查找的对象填入下面的指针类型变量中
	// dep为期望状态不是指针，found为实际集群的状态可以实时反应出来
	found := &appsv1.Deployment{}
	// 在集群中查找dep对象
	// type ObjectKey types.NamespacedName
	// Object 需要是一个指针类型
	// 找到了就为空
	// 每次更新deployment出现更新就会触发这个操作
	err = r.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, found)

	if err != nil {
		// 2 调用 client/Writer 接口来往k8s里面创建资源
		err = r.Create(context.TODO(), dep)
		if err != nil {
			clog.Error(err, "deployment create not ok")
		} else {
			//r.Recorder.Event(dep, "Normal", "GetRes", "deployment create ok !")
			clog.Info("deployment create ok", "deployment-name", dep.Name)
			MacBook.Status.Mod = dep.Name
			// 关键代码 更新status
			if err := r.Status().Update(ctx, MacBook); err != nil {
				clog.Error(err, "MacBook.Status.Mod update fail !")
			} else {
				clog.Info("MacBook.Status.Mod update okokok !")
			}
		}
	} else {
		//clog.Info("找到了 deployment", "lable", dep.Spec.Template.Spec.Containers[0].Name)
		clog.Info("找到了 deployment", "Annotations", found.Annotations)
	}

	depList := &appsv1.DeploymentList{}
	if err := r.List(ctx, depList, client.InNamespace(req.Namespace), client.MatchingFields{nsKey: req.Namespace}); err != nil {
		return ctrl.Result{}, err
	}

	clog.Info("获取到了某个ns的deployment列表", "delListLen", len(depList.Items))

	return ctrl.Result{}, nil

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

// predicate 使用 https://sdk.operatorframework.io/docs/building-operators/golang/references/event-filtering/
func onlyReconcilerDeploymentLable() predicate.Predicate {
	return predicate.Funcs{
		// 这些函数中返回值为 true 就会执行响应的事件handler
		// 下面的逻辑表示 在更新事件中，只有ns为 lrq 的资源对象才会触发更新事件handler
		UpdateFunc: func(event event.UpdateEvent) bool {
			return event.ObjectNew.GetNamespace() == "lr"
		},
		CreateFunc: func(createEvent event.CreateEvent) bool {
			return createEvent.Object.GetNamespace() == "lr"
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return deleteEvent.Object.GetNamespace() == "lr"
		},
	}
}

var nsKey = "byNs"

//var nsKey = ".metadata.namespace"

// SetupWithManager sets up the controller with the Manager.
func (r *MacBookReconciler) SetupWithManager(mgr ctrl.Manager) error {

	/*
		加快搜索，增加索引
		https://github.com/kubernetes-sigs/kubebuilder/issues/1422
		这个函数相当于自己为某个资源对象创建了一个数据的维度，这个维度字段可以使用客户端的 client.MatchingFields 方法来选择使用
		这里形成的是个"倒排索引"
		资源对象就是一个个的文档，根据某个文档中的字段进行索引给这个索引起个名字就是 nskey ，
		倒排索引中key 就是 对象中的某一个key的value 比如，value 就是 该文档
		byNs
		lr  deployment1，deployment2,deployment3
		defalut deployment5
	*/
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &appsv1.Deployment{}, nsKey, func(rawObj client.Object) []string {

		deployment := rawObj.(*appsv1.Deployment)
		ns := deployment.GetNamespace()

		return []string{ns}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		// for指定需要监听的资源 基于watch实现
		// Watches(&source.Kind{Type: apiType}, &handler.EnqueueRequestForObject{})
		// builder.WithPredicates(predicate.GenerationChangedPredicate{}) 忽略status字段更新的调协操作
		For(&mockv1beta1.MacBook{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		// Owns 指定监听crd的子资源,第二个字段是过滤器，针对不同的事件采取特定的过滤策略
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.Deployment{}, builder.WithPredicates(onlyReconcilerDeploymentLable())).
		Complete(r)
}
