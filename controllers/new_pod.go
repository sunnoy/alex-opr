/*
 *@Description
 *@author          lirui
 *@create          2021-04-04 12:21
 */
package controllers

import (
	mockv1beta1 "alex-opr/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

func newCreatePod(ins *mockv1beta1.MacBook) *corev1.Pod {
	labels := map[string]string{
		"app": ins.Name,
	}
	var a int64 = 0

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ins.Name + "-pod",
			Namespace: ins.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				// 第一个容器
				{
					Name:    "busybox",
					Image:   "prdharbor.xylink.com/internal/busybox:1.30",
					Command: strings.Split("sleep 100000", " "),
				},
			},
			ImagePullSecrets: []corev1.LocalObjectReference{
				{
					Name: "harbor-xylink-com",
				},
			},
			RestartPolicy:                 corev1.RestartPolicyOnFailure,
			TerminationGracePeriodSeconds: &a,
		},
	}

}
