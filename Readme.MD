# Kubernetes Operator

1) This operator is used to set up this <a href="https://github.com/Ashwin901/Social-Book-Server">application</a>.
2) The application includes a NodeJs server and a MongoDB database.
3) Docker image of the server can be found <a href="https://hub.docker.com/repository/docker/ashwin901/social-book-server">here</a>.

## Components

<ul>
    <li>
        Custom Resource
    </li>
    <li>
        Controller
    </li>
</ul>

### Custom Resource

<ul>
    <li><a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/manifests/operators_socialbooks.yaml">CRD</a></li>
    <li>Kind: <b>SocialBook</b>, Group: <b>ashwin901.operators</b></li>
    <li>Example for the Custom Resource can be found <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/manifests/example1.yml">here</a></li>
</ul>

### Controller

1. Controller code can be found <a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/controller.go">here</a>.
2. When a SocialBook CR is created the custom controller sets up the following resources: 
    <ul>    
        <li><a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/configmap.go">Config Map</a></li>
        <li><a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/persistentvolume.go">Persistent Volume</a></li>
        <li><a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/persistentvolume.go">Persistent Volume Claim</a></li>
        <li><a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/deployment.go">Deployment - MongoDB and SocialBook</a></li>
        <li><a href="https://github.com/Ashwin901/K8s-Operator-SocialBook/blob/master/controller/service.go">Services</a></li>
    </ul>

3. If any of the above mentioned resource is updated/deleted then the custom controller will detect the change and try to get it back to the desired state.

4. If a particular SocialBook resource is deleted then all the resources setup for it will also be deleted. This is done with the help of owner reference.

## Tools

1. <a href="https://github.com/kubernetes/code-generator">Code Generator</a> - To generate code for clientset, informers and lister
2. <a href="https://book.kubebuilder.io/reference/controller-gen.html">Controller Gen</a> - To generate manifests for CRD


