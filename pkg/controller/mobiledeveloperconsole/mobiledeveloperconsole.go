package mobiledeveloperconsole

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/aerogear/mobile-developer-console-operator/pkg/util"
	"github.com/aerogear/mobile-developer-console-operator/version"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/intstr"

	mdcv1alpha1 "github.com/aerogear/mobile-developer-console-operator/pkg/apis/mdc/v1alpha1"
	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	openshiftappsv1 "github.com/openshift/api/apps/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newMDCServiceAccount(cr *mdcv1alpha1.MobileDeveloperConsole) (*corev1.ServiceAccount, error) {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Annotations: map[string]string{
				"serviceaccounts.openshift.io/oauth-redirectreference.mdc": fmt.Sprintf("{\"kind\":\"OAuthRedirectReference\",\"apiVersion\":\"v1\",\"reference\":{\"kind\":\"Route\",\"name\":\"%s-mdc-proxy\"}}", cr.Name),
			},
		},
	}, nil
}

func newOauthProxyService(cr *mdcv1alpha1.MobileDeveloperConsole) (*corev1.Service, error) {
	return &corev1.Service{
		ObjectMeta: util.ObjectMeta(&cr.ObjectMeta, "mdc-proxy"),
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":     cr.Name,
				"service": "mdc",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "web",
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 4180,
					},
				},
			},
		},
	}, nil
}

func newMDCService(cr *mdcv1alpha1.MobileDeveloperConsole) (*corev1.Service, error) {
	serviceObjectMeta := util.ObjectMeta(&cr.ObjectMeta, "mdc")
	serviceObjectMeta.Labels["mobile"] = "enabled"
	serviceObjectMeta.Labels["internal"] = "mdc"

	return &corev1.Service{
		ObjectMeta: serviceObjectMeta,
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app":     cr.Name,
				"service": "mdc",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "web",
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 4000,
					},
				},
			},
		},
	}, nil
}

func newOauthProxyRoute(cr *mdcv1alpha1.MobileDeveloperConsole) (*routev1.Route, error) {
	return &routev1.Route{
		ObjectMeta: util.ObjectMeta(&cr.ObjectMeta, "mdc-proxy"),
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: fmt.Sprintf("%s-%s", cr.Name, "mdc-proxy"),
			},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyNone,
			},
		},
	}, nil
}

func newOauthProxyImageStream(cr *mdcv1alpha1.MobileDeveloperConsole) (*imagev1.ImageStream, error) {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      cfg.OauthProxyImageStreamName,
			Labels:    util.Labels(&cr.ObjectMeta, cfg.OauthProxyImageStreamName),
		},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				{
					Name: cfg.OauthProxyImageStreamTag,
					From: &corev1.ObjectReference{
						Kind: "DockerImage",
						Name: cfg.OauthProxyImageStreamInitialImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Scheduled: false,
					},
				},
			},
		},
	}, nil
}

func newMDCImageStream(cr *mdcv1alpha1.MobileDeveloperConsole) (*imagev1.ImageStream, error) {
	return &imagev1.ImageStream{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      cfg.MDCImageStreamName,
			Labels:    util.Labels(&cr.ObjectMeta, cfg.MDCImageStreamName),
		},
		Spec: imagev1.ImageStreamSpec{
			Tags: []imagev1.TagReference{
				{
					Name: cfg.MDCImageStreamTag,
					From: &corev1.ObjectReference{
						Kind: "DockerImage",
						Name: cfg.MDCImageStreamInitialImage,
					},
					ImportPolicy: imagev1.TagImportPolicy{
						Scheduled: false,
					},
				},
			},
		},
	}, nil
}

func newMDCDeploymentConfig(cr *mdcv1alpha1.MobileDeveloperConsole) (*openshiftappsv1.DeploymentConfig, error) {
	labels := map[string]string{
		"app":     cr.Name,
		"service": "mdc",
	}

	cookieSecret, err := util.GeneratePassword()
	if err != nil {
		return nil, errors.Wrap(err, "error generating cookie secret")
	}

	return &openshiftappsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: openshiftappsv1.DeploymentConfigSpec{
			Replicas: 1,
			Selector: labels,
			Triggers: openshiftappsv1.DeploymentTriggerPolicies{
				openshiftappsv1.DeploymentTriggerPolicy{
					Type: openshiftappsv1.DeploymentTriggerOnConfigChange,
				},
				openshiftappsv1.DeploymentTriggerPolicy{
					Type: openshiftappsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &openshiftappsv1.DeploymentTriggerImageChangeParams{
						Automatic:      true,
						ContainerNames: []string{cfg.MDCContainerName},
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: cfg.MDCImageStreamName + ":" + cfg.MDCImageStreamTag,
						},
					},
				},
				openshiftappsv1.DeploymentTriggerPolicy{
					Type: openshiftappsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &openshiftappsv1.DeploymentTriggerImageChangeParams{
						Automatic:      true,
						ContainerNames: []string{cfg.OauthProxyContainerName},
						From: corev1.ObjectReference{
							Kind: "ImageStreamTag",
							Name: cfg.OauthProxyImageStreamName + ":" + cfg.OauthProxyImageStreamTag,
						},
					},
				},
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: cr.Name,
					Containers: []corev1.Container{
						{
							Name:            cfg.MDCContainerName,
							Image:           cfg.MDCImageStreamName + ":" + cfg.MDCImageStreamTag,
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name:  "NAMESPACE",
									Value: cr.Namespace,
								},
								{
									Name:  "NODE_ENV",
									Value: "production",
								},
								{
									Name:  "OPENSHIFT_HOST",
									Value: cfg.OpenShiftHost,
								},
								{
									Name:  "IDM_DOCUMENTATION_URL",
									Value: cfg.IdentityManagementDocumentationURL,
								},
								{
									Name:  "UPS_DOCUMENTATION_URL",
									Value: cfg.UnifiedPushDocumentationURL,
								},
								{
									Name:  "SYNC_DOCUMENTATION_URL",
									Value: cfg.DataSyncDocumentationURL,
								},
								{
									Name:  "MSS_DOCUMENTATION_URL",
									Value: cfg.MobileSecurityDocumentationURL,
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          cfg.MDCContainerName,
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 4000,
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("256Mi"),
									corev1.ResourceCPU:    resource.MustParse("100m"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("128Mi"),
									corev1.ResourceCPU:    resource.MustParse("50m"),
								},
							},
						},
						{
							Name:            cfg.OauthProxyContainerName,
							Image:           cfg.OauthProxyImageStreamName + ":" + cfg.OauthProxyImageStreamTag,
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									Name:          "public",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 4180,
								},
							},
							Args: []string{
								"--provider=openshift",
								fmt.Sprintf("--client-id=%s", cr.Spec.OAuthClientId),
								fmt.Sprintf("--client-secret=%s", cr.Spec.OAuthClientSecret),
								"--upstream=http://localhost:4000",
								"--http-address=0.0.0.0:4180",
								"--https-address=",
								fmt.Sprintf("--cookie-secret=%s", cookieSecret),
								"--cookie-httponly=false", // we kill the possibility to run MDC on a http route
								"--pass-access-token=true",
								"--scope=user:full",
								"--bypass-auth-for=/about",
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("64Mi"),
									corev1.ResourceCPU:    resource.MustParse("20m"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceMemory: resource.MustParse("32Mi"),
									corev1.ResourceCPU:    resource.MustParse("10m"),
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func newMobileClientAdminRoleBinding(cr *mdcv1alpha1.MobileDeveloperConsole) (*rbacv1.RoleBinding, error) {
	name := cr.Name + "-mobileclient-admin"
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      name,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     name,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      rbacv1.ServiceAccountKind,
				Name:      cr.Name,
				Namespace: cr.Namespace,
			},
		},
	}, nil
}

func newMobileClientAdminRole(cr *mdcv1alpha1.MobileDeveloperConsole) (*rbacv1.Role, error) {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      cr.Name + "-mobileclient-admin",
		},
		Rules: []rbacv1.PolicyRule{
			rbacv1.PolicyRule{
				APIGroups: []string{""},
				Resources: []string{"secrets", "configmaps"},
				Verbs:     []string{"get", "list", "watch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"mdc.aerogear.org"},
				Resources: []string{"mobileclients"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"push.aerogear.org"},
				Resources: []string{"pushapplications", "androidvariants", "iosvariants", "webpushvariants"},
				Verbs:     []string{"get", "list", "watch"},
			},
			rbacv1.PolicyRule{
				APIGroups: []string{"mobile-security-service.aerogear.org"},
				Resources: []string{"mobilesecurityserviceapps"},
				Verbs:     []string{"get", "list", "watch"},
			},
		},
	}, nil
}

func newMDCServiceMonitor(cr *mdcv1alpha1.MobileDeveloperConsole) (*monitoringv1.ServiceMonitor, error) {
	labels := map[string]string{
		"monitoring-key": "middleware",
		"name":           "mobile-developer-console",
	}
	matchLabels := map[string]string{
		"internal": "mdc",
	}
	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      "mobile-developer-console",
			Labels:    labels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			Endpoints: []monitoringv1.Endpoint{
				{
					Path: "/metrics",
					Port: "web",
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
		},
	}, nil
}

func newMDCPrometheusRule(cr *mdcv1alpha1.MobileDeveloperConsole) (*monitoringv1.PrometheusRule, error) {
	labels := map[string]string{
		"monitoring-key": "middleware",
		"prometheus":     "application-monitoring",
		"role":           "alert-rules",
	}
	critical := map[string]string{
		"severity": "critical",
	}
	warning := map[string]string{
		"severity": "warning",
	}
	sop_url := fmt.Sprintf("https://github.com/aerogear/mobile-developer-console-operator/blob/%s/SOP/SOP-mdc.adoc", version.Version)
	mdc_info := "For more information see on the MDC at https://github.com/aerogear/mobile-developer-console"
	mdcContainerDownAnnotations := map[string]string{
		"description": "The MDC has been down for more than 5 minutes.",
		"summary":     fmt.Sprintf("The mobile-developer-console is down. %s", mdc_info),
		"sop_url":     sop_url,
	}
	mdcDownAnnotations := map[string]string{
		"description": "The MDC admin console has been down for more than 5 minutes.",
		"summary":     fmt.Sprintf("The mobile-developer-console admin console endpoint has been unavailable for more that 5 minutes. %s", mdc_info),
		"sop_url":     sop_url,
	}
	mdcPodCPUHighAnnotations := map[string]string{
		"description": "The MDC pod has been at 90% CPU usage for more than 5 minutes",
		"summary":     fmt.Sprintf("The mobile-developer-console is reporting high cpu usage for more that 5 minutes. %s", mdc_info),
		"sop_url":     sop_url,
	}
	mdcPodMemHighAnnotations := map[string]string{
		"description": "The MDC pod has been at 90% memory usage for more than 5 minutes",
		"summary":     fmt.Sprintf("The mobile-developer-console is reporting high memory usage for more that 5 minutes. %s", mdc_info),
		"sop_url":     sop_url,
	}
	serviceObjectMeta := util.ObjectMeta(&cr.ObjectMeta, "mdc")
	container := "mdc"
	return &monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      "mdc-monitoring",
			Labels:    labels,
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: "general.rules",
					Rules: []monitoringv1.Rule{
						{
							Alert: "MobileDeveloperConsoleContainerDown",
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: fmt.Sprintf("absent(kube_pod_container_status_running{namespace=\"%s\",container=\"%s\"}>=1)", cr.Namespace, container),
							},
							For:         "5m",
							Labels:      critical,
							Annotations: mdcContainerDownAnnotations,
						},
						{
							Alert: "MobileDeveloperConsoleDown",
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: fmt.Sprintf("absent(kube_endpoint_address_available{endpoint=\"%s\"} >= 1)", serviceObjectMeta.Name),
							},
							For:         "5m",
							Labels:      critical,
							Annotations: mdcDownAnnotations,
						},
						{
							Alert: "MobileDeveloperConsolePodCPUHigh",
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: fmt.Sprintf("(rate(process_cpu_seconds_total{job='%s'}[1m])) > (((kube_pod_container_resource_limits_cpu_cores{namespace='%s',container='%s'})/100)*90)", serviceObjectMeta.Name, cr.Namespace, container),
							},
							For:         "5m",
							Labels:      warning,
							Annotations: mdcPodCPUHighAnnotations,
						},
						{
							Alert: "MobileDeveloperConsolePodMemoryHigh",
							Expr: intstr.IntOrString{
								Type:   intstr.String,
								StrVal: fmt.Sprintf("(process_resident_memory_bytes{job='%s'}) > (((kube_pod_container_resource_limits_memory_bytes{namespace='%s',container='%s'})/100)*90)", serviceObjectMeta.Name, cr.Namespace, container),
							},
							For:         "5m",
							Labels:      warning,
							Annotations: mdcPodMemHighAnnotations,
						},
					},
				},
			},
		},
	}, nil
}
