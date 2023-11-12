# Testing Kubernetes Controllers with the e2e-framework

This example shows you how to create end-to-end tests to test Kubernetes controller using the [CronJob controller](https://book.kubebuilder.io/cronjob-tutorial/cronjob-tutorial) example that comes with the The Kubernetes-SIGs [Kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) project. It is based on blog post about the subject on [Medium](https://medium.com/programming-kubernetes/testing-kubernetes-controllers-with-the-e2e-framework-fac232843dc6).

Find the example end-to-end test source code in [./testdata/e2e-test](./testdata/e2e-test/).

### Infrastructure setup

First, we will use the `Environment.Setup` method, in `TestMain`, to create a new Kind cluster and install a namespace on the cluster:

```go
var (
   testEnv env.Environment
   namespace = "cronjob"
)

func TestMain(m *testing.M) {
   testEnv = env.New()
   kindClusterName := "kind-test"
   kindCluster := kind.NewCluster(kindClusterName)

   testEnv.Setup(
     envfuncs.CreateCluster(kindCluster, kindClusterName),
     envfuncs.CreateNamespace(namespace),
     ...
   )
}
```

Next, we will install prometheus and cert-manager and wait for the cert manager deployment to be ready before continuing:

```go
var (
...
   certmgrVer = "v1.13.1"
   certMgrUrl = fmt.Sprintf("https://github.com/jetstack/cert-manager/releases/download/%s/cert-manager.yaml", certmgrVer)
   
   promVer    = "v0.60.0"
   promUrl    = fmt.Sprintf("https://github.com/prometheus-operator/prometheus-operator/releases/download/%s/bundle.yaml", promVer)
)

func TestMain(m *testing.M) {
...
  testEnv.Setup(  
    func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
       // install prometheus
       if p := utils.RunCommand(
          fmt.Sprintf("kubectl apply -f %s --server-side", promUrl),
       ); p.Err() != nil {
          return ctx, p.Err()
       }


       // install cert-manager
       if p := utils.RunCommand(
          fmt.Sprintf("kubectl apply -f %s", certMgrUrl),
       ); p.Err() != nil {
          return ctx, p.Err()
       }

       // wait for certmgr deployment to be ready
       client := cfg.Client()
       if err := wait.For(
          conditions.New(client.Resources()).
             DeploymentAvailable("cert-manager-webhook", "cert-manager"),
          wait.WithTimeout(5*time.Minute),
          wait.WithInterval(10*time.Second),
       ); err != nil {
          return ctx, err
       }
       return ctx, nil
    },
   )
}
```

The next function installs `kustomize` and `controller-gen` needed to generate the required source and configuration files for the controller:

```go
var (
...
   kustomizeVer = "v5.1.1"
   ctrlgenVer   = "v0.13.0"
)

func TestMain(m *testing.M) {
  ...
  testEnv.Setup(
    ...
    func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
       // install kubstomize binary
       if p := utils.RunCommand(
          fmt.Sprintf("go install sigs.k8s.io/kustomize/kustomize/v5@%s", kustomizeVer),
       ); p.Err() != nil {
          return ctx, p.Err()
       }
       
       // install controller-gen binary
       if p := utils.RunCommand(
          fmt.Sprintf("go install sigs.k8s.io/controller-tools/cmd/controller-gen@%s", ctrlgenVer),
       ); p.Err() != nil {
          return ctx, p.Err()
       }
       return ctx, nil
    },
  )
}
```

Using the binaries installed in the previous step, we can now define a step function to:

* Generate the configuration and source files for the controller
* Build and deploy the controller container image to Docker
* Load the built image into Kind
* Deploy the controller components to the local Kind cluster

```go
func TestMain(m *testing.M) {
  ...
  testEnv.Setup(
    ...
    func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
       // gen manifest files
       if p := utils.RunCommand(
          `controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases`,
       ); p.Err() != nil {
          return ctx, p.Err()
       }

       // gen api objects
       if p := utils.RunCommand(
          `controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."`,
       ); p.Err() != nil {
          return ctx, p.Err()
       }

       // Build the docker image
       if p := utils.RunCommand(
          fmt.Sprintf("docker build -t %s .", dockerImage),
       ); p.Err() != nil {
          return ctx, p.Err()
       }

       // Load the docker image into kind
       if err := kindCluster.LoadImage(ctx, dockerImage); err != nil {
          return ctx, err
       }

       // Deploy the controller components
       if p := utils.RunCommand(
          `bash -c "kustomize build config/default | kubectl apply --server-side -f -"`,
       ); p.Err() != nil {
           return ctx, p.Err()
       }

       // wait for the controller deployment to be ready
       client := cfg.Client()
       if err := wait.For(
          conditions.New(client.Resources()).
             DeploymentAvailable("cronjob-controller-manager", "cronjob-system"),
          wait.WithTimeout(3*time.Minute),
          wait.WithInterval(10*time.Second),
       ); err != nil {
          return ctx, err
       }
       return ctx, nil
    },
  )
}
```

Lastly, let’s define some clean up steps to tear down the infrastructure after all tests are completed:

```go
testEnv.Finish(
   // remove cluster components
   func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
      utils.RunCommand(fmt.Sprintf("kubectl delete -f %s", promUrl))
      utils.RunCommand(fmt.Sprintf("kubectl delete -f %s", certMgrUrl))
      utils.RunCommand(`bash -c "kustomize build config/default | kubectl delete -f -"`)
      return ctx, nil
   },
   // delete namespace and destroy cluster
   envfuncs.DeleteNamespace(namespace),
   envfuncs.DestroyCluster(kindClusterName),
 )
 ```

### The test function
Now that we have setup an infrastructure with a Kubernetes cluster, we can use a Go test function to define unit tests for the controller.

But first, let’s define a `CronJob` struct literal that we will use in the tests:

```go
var (
   cronjobName = "cronjob-controller"

   cronjob = &cronjobV1.CronJob{
      TypeMeta: metaV1.TypeMeta{
         APIVersion: "batch.tutorial.kubebuilder.io/v1",
         Kind:       "CronJob",
      },
      ObjectMeta: metaV1.ObjectMeta{
         Name:      cronjobName,
         Namespace: namespace,
      },
      Spec: cronjobV1.CronJobSpec{
         Schedule: "1 * * * *",
         JobTemplate: batchV1.JobTemplateSpec{
            Spec: batchV1.JobSpec{
               Template: coreV1.PodTemplateSpec{
                  Spec: coreV1.PodSpec{
                     Containers: []coreV1.Container{{
                        Name:  "test-container",
                        Image: "test-image",
                     }},
                     RestartPolicy: coreV1.RestartPolicyOnFailure,
                  },
               },
            },
         },
      },
   }
)
```

Next, let's create a `Feature.Setup` to setup up a `Watcher` to watch for job pods created by our controller and put them on channel `podCreationSig` to be tested later:

```go
func TestCron(t *testing.T) {
   podCreationSig := make(chan *coreV1.Pod)

   feature := features.New("Cronjob Controller")
   feature.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
      client := cfg.Client()

   // setup watcher for creation of job v1.Pods by cronjob-controller
   if err := client.Resources(namespace).
         Watch(&coreV1.PodList{}).
         WithAddFunc(func(obj interface{}) {
            pod := obj.(*coreV1.Pod)
           if strings.HasPrefix(pod.Name, "cronjob-controller") {
              podCreationSig <- pod
           }
   }).Start(ctx); err != nil {
      t.Fatal(err)
   }
      return ctx
   })
   ...
}
```

Next, lets add an assessment to ensure the CronJob CRD is deployed on the cluster :

```go
func TestCron(t *testing.T) {
   ...
   feature.Assess("CRD installed", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
      client := cfg.Client()
      apiextensionsV1.AddToScheme(client.Resources().GetScheme())
      name := "cronjobs.batch.tutorial.kubebuilder.io"
      var crd apiextensionsV1.CustomResourceDefinition
      if err := client.Resources().Get(ctx, name, "", &crd); err != nil {
         t.Fatalf("CRD not found: %s", err)
      }

      if crd.Spec.Group != "batch.tutorial.kubebuilder.io" {
         t.Fatalf("Cronjob CRD has unexpected group: %s", crd.Spec.Group)
      }
      return ctx
   })
}
```

In the next assessment, we will create a new instance of a `CronJob` in the cluster and use the `wait` package to construct a dierctive to wait for API object to be created:

```go
func TestCron(t *testing.T) {
   ...
   feature.Assess("Cronjob creation", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
      client := cfg.Client()
      cronjobV1.AddToScheme(client.Resources(namespace).GetScheme())

      if err := client.Resources().Create(ctx, cronjob); err != nil {
         t.Fatalf("Failed to crete cronjob: %s", err)
      }

      // wait for resource to be created
      if err := wait.For(
         conditions.New(client.Resources()).
            ResourceMatch(cronjob, func(object k8s.Object) bool {
               return true
            },
         ),
         wait.WithTimeout(3*time.Minute),
         wait.WithInterval(30*time.Second),
      ); err != nil {
         t.Fatal(err)
      }

      return ctx
   })
}
```

When the controller creates a new `CronJob` custom resource, it should also create a pod for the job. The following assessment uses the `podCreationSig` channel, declared earlier, to block until it receives an instance of the created pod or fails if it does not within the specified time:

```go
func TestCron(t *testing.T) {
   ...
   feature.Assess("Watcher received pod job", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
      select {
         case <-time.After(30 * time.Second):
            t.Error("Timed out wating for job pod creation by cronjob contoller")
         case pod := <-podCreationSig:
            t.Log("Pod created by cronjob-controller")
            refname := pod.GetOwnerReferences()[0].Name
            if !strings.HasPrefix(refname, cronjobName) {
               t.Fatalf("Job pod has unexpected owner ref: %#v", refname)
            }
      }
      return ctx
   })
}
```

### Running the test
Next, use the `go test` command to run the test. Go will automtomatically pull down all required packages and start the test:

```
$> cd ./testdata
$> go test -v ./e2e-test

go: downloading sigs.k8s.io/e2e-framework v0.3.0
go: downloading k8s.io/api v0.28.3
go: downloading k8s.io/apimachinery v0.28.3
...
```