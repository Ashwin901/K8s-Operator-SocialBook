# Kubernetes Operator

1) This operator is used to set up this <a href="https://github.com/Ashwin901/Social-Book-Server">application</a>.
2) The application includes a NodeJs server and a MongoDB database.
3) Docker image of the server can be found <a href="https://hub.docker.com/repository/docker/ashwin901/social-book-server">here</a>.

## Components

1. Custom Resource
2. Controller

### Custom Resource

1. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/manifests/operators_socialbooks.yaml">CRD</a>
2. Group: <b>ashwin901.operators</b> Version: <b>v1alpha1</b> Kind: <b>SocialBook</b>
3. Example for the Custom Resource can be found <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/manifests/example1.yml">here</a>


### Controller

1. Controller code can be found <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/controller.go">here</a>.
2. When a SocialBook CR is created the custom controller sets up the following resources: 
        1. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/configmap.go">Config Map</a>
        2. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/persistentvolume.go">Persistent Volume</a>
        3. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/persistentvolume.go">Persistent Volume Claim</a>
        4. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/deployment.go">Deployment - MongoDB and SocialBook</a>
        5. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/service.go">Services</a>
        6. <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/networkPolicy.go">Network Policy</a>
    
    

3. If any of the above mentioned resource is updated/deleted then the custom controller will detect the change and try to get it back to the desired state.

4. If a particular SocialBook resource is deleted then all the resources setup for it will also be deleted. This is done with the help of owner reference.

## Tools

1. <a href="https://github.com/kubernetes/code-generator">Code Generator</a> - To generate code for clientset, informers and lister
2. <a href="https://book.kubebuilder.io/reference/controller-gen.html">Controller Gen</a> - To generate manifests for CRD


