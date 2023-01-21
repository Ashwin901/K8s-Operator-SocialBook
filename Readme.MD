# Kubernetes Operator

### Description
This operator is used to set up this <a href="https://github.com/Ashwin901/Social-Book-Server">application</a>. The application includes a Nodejs server and a MongoDB database. Docker image of the server can be found <a href="https://hub.docker.com/repository/docker/ashwin901/social-book-server">here</a>. <br/>
When a new `SocialBook` custom resource is created the custom controller will create a `MongoDB` deployment, a corresponding service for it, Persistent Volume and Persistent Volume Claim. It will also create a `SocialBook` deployment and an external service so that it can be accessed outside the cluster.<br/>
Apart from this it will also create a Config Map and Network policies for both MongoDB and SocialBook pods. The number of replicas and other information can be passed in the spec of the custom resource.

> Note: For the network policies to work, a network plugin should already be installed on the cluster.

### Usage

#### Outside the cluster
1. Clone the repo - `git clone https://github.com/Ashwin901/K8s-Operator-SocialBook.git`
2. Run `kubectl apply -f manifests/operators_socialbooks.yaml` to install the CRD. 
3. Run `go build` from the parent directory.
4. Run `./social-book-operator --config <kube config file path>` (In linux systems usually kube config file path is /home/username/.kube/config).

#### Inside the cluster
1. To install the operator inside the cluster we can use the docker image of the <a href="https://hub.docker.com/repository/docker/ashwin901/social-book-operator">operator</a>.
2. Install the CRD by using the following <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/manifests/operators_socialbooks.yaml">file</a>.
3. Copy the files from this <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/tree/master/manifests/install">directory</a>. 
4. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/manifests/install/rbac.yml">rbac.yml</a> file conists of a service account, cluster role and cluster role binding. This basically gives the operator permissions to access Kubernetes resources. Run `kubectl apply -f rbac.yml`.
5. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/manifests/install/deployment.yml">deployment.yml</a> creates a deployment for the operator docker image. Run `kubectl apply -f deployment.yml`

Now you can test the operator by creating a new SocialBook custom resource. You can use this <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/manifests/example1.yml">example</a>. Run `kubectl apply -f example1.yml`. 
Once the custom resource is created check the `dev` namespace(in the above example `dev` namespace is used but you can use any namespace) if all the resources are created.

#### Accessing the app
If you are using minikube use the following command: `minikube service -n dev socialbook1` (`socialbook1` -  name used in the above example)


### Components

1. Custom Resource
2. Controller

#### Custom Resource

1. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/manifests/operators_socialbooks.yaml">CRD</a>
2. Group: <b>ashwin901.operators</b> Version: <b>v1alpha1</b> Kind: <b>SocialBook</b>
3. Example for the Custom Resource can be found <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/manifests/example1.yml">here</a>


#### Controller

1. Controller code can be found <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/controller.go">here</a>.
2. When a SocialBook CR is created the custom controller sets up the following resources: <br/>
        1. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/configmap.go">Config Map</a><br/>
        2. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/persistentvolume.go">Persistent Volume</a><br/>
        3. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/persistentvolume.go">Persistent Volume Claim</a><br/>
        4. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/deployment.go">Deployment - MongoDB and SocialBook</a><br/>
        5. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/service.go">Services</a><br/>
        6. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/networkPolicy.go">Network Policy</a> - Ensures that the `MongoDB` pod only accepts requests from `SocialBook` pods(ingress) and `SocialBook` pods can only make requests to `MongoDB` pods(egress).
3. If any of the above mentioned resource is updated/deleted then the custom controller will detect the change and try to get it back to the desired state.

4. If a particular SocialBook resource is deleted then all the resources setup for it will also be deleted. This is done with the help of owner reference.

### Tools

1. <a href="https://github.com/kubernetes/code-generator">Code Generator</a> - To generate code for clientset, informers and lister
2. <a href="https://book.kubebuilder.io/reference/controller-gen.html">Controller Gen</a> - To generate manifests for CRD


