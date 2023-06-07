### Service Functioning
This repository is a fork of the open source [krakend](https://github.com/krakendio/krakend-ce) API gateway.
It is used in Unacademy for the following use cases
1. Routing requests transparently to multiple upstream microservices and act as a single entrypoint for our services. The routing config can be viewed in [krakend-config](https://github.com/Unacademy/krakend-config) repo.
2. Validating the auth token passed in the request to offload authentication from upstream services.

### Local Setup
Install gvm (go version manager) to setup go virtual env, if you already have a virtual env, you can use that as well

#### steps to install gvm and use activate go virtual env
1. install gvm - `bash < <(curl -s -S -L https://raw.githubusercontent.com/moovweb/gvm/master/binscripts/gvm-installer)`
2. install go1.16 version `gvm install go1.17`
3. use the installed go version `gvm use go1.17`

#### Setup and build krakend-ce
1. Install dependencies - `go get .`
2. build binary - 
    ```
      go build \
    -ldflags="-X github.com/devopsfaith/krakend/core.KrakendVersion=1.4.2 \
    -X github.com/devopsfaith/krakend/core.KrakendHeaderName=X-Gateway \
    -X github.com/devopsfaith/krakend/transport/http/server.CompleteResponseHeaderName=X-Gateway-Completed" \
    -o krakend ./cmd/krakend-ce
    ```
3. install runtime plugins
    - clone `krakend-plugins` repository and cd into the directory
    - build plugins in plugin mode using - `make all`
    - copy plugins into the plugins directory `mkdir plugins && cp <krakend-plugins-dir>/bin/*.so plugins/`
4. copy config as per your requirement, to get gamma config, you can copy it from the krakend gamma pipeline, or if you want to test, you can use `examples/simple_config.json` to tet.
5. copy the env variables from gamma vault and export that variables in your terminal. Comment out variables `FC_ENABLE FC_OUT FC_PARTIALS FC_SETTINGS FC_TEMPLATES` since you are directly copying the completely built config.
6. run `./krakend run -c <path-to-config-file>`
7. if you want to setup debugger in vscode, I have attached launch.json file, you can directly run the debugger directly using that, you would have to just update the env key from env from vault.


#### Docker setup
You can use the following Dockerfile to setup docker container for krakend.
Prequisites
  - Docker engine
  - AWS prod account configs
  - ECR Access for your IAM role

```
FROM 838337956332.dkr.ecr.ap-southeast-1.amazonaws.com/dockerhub:golang_1.17.6

ARG github_token
RUN git config --global url."https://${github_token}@github.com/".insteadOf "https://github.com/"
ENV github_token ""

WORKDIR /krakend

COPY . .
RUN go get .; go build \
  -ldflags="-X github.com/devopsfaith/krakend/core.KrakendVersion=1.4.2 \
  -X github.com/devopsfaith/krakend/core.KrakendHeaderName=X-Gateway \
  -X github.com/devopsfaith/krakend/transport/http/server.CompleteResponseHeaderName=X-Gateway-Completed" \
  -o krakend ./cmd/krakend-ce

# pull krakend-plugins repository
ARG KRAKEND_PLUGINS_BRANCH
RUN git clone https://github.com/Unacademy/krakend-plugins.git -b $KRAKEND_PLUGINS_BRANCH
RUN cd krakend-plugins; make all
RUN mkdir plugins; cp krakend-plugins/bin/*.so plugins/

# copy krakend config files

EXPOSE 8080

ENTRYPOINT [".", ".env.sh"]

CMD ./krakend run -c <config-file-path>
```



### Prod Deployment
To deploy to production, merge your changes in the master branch, and then trigger the following jenkins [pipeline](https://jenkins.unacademydev.com/job/sphere/job/krakend/job/master/)

### Metrics to Monitor
Krakend is attached as a target group in the main Unacademy API load balancer, so to view the metrics for krakend, you need to view the metrics for the krakend tg i.e `k8s-krakend-tg`

Alarms
1. [5xx High](https://ap-southeast-1.console.aws.amazon.com/cloudwatch/home?region=ap-southeast-1#alarmsV2:alarm/krakend_5XX_critical?~(search~'krakend)), threshold 350 in 5mins

Metrics
You should monitor the following metrics in the `ApplicationELB, Per AppELB, per TG Metrics` namespace, [link](https://ap-southeast-1.console.aws.amazon.com/cloudwatch/home?region=ap-southeast-1#metricsV2:graph=~(view~'timeSeries~stacked~false~region~'ap-southeast-1~stat~'Sum~period~300);query=~'*7bAWS*2fApplicationELB*2cLoadBalancer*2cTargetGroup*7d*20k8s-krakend-tg)
1. RequestCount
2. HTTPCode_Target_4XX_Count
3. HTTPCode_Target_2XX_Count
4. TargetResponseTime
5. HTTPCode_Target_5XX_Count

### Debugging
To debug the service in production, you can view the logs for your service on Opensearch in the index pattern [p-krakend-20*](https://vpc-logs-001-b5nny2i64do3jtxlqsojexazma.ap-southeast-1.es.amazonaws.com/_dashboards/app/discover#/?_g=(filters:!(),query:(language:kuery,query:''),refreshInterval:(pause:!t,value:0),time:(from:now-15h,to:now))&_a=(columns:!(_source),filters:!(),index:'7cd38c30-bcd8-11ed-8128-897093931142',interval:auto,query:(language:kuery,query:''),sort:!()))

### Contact Information
[Sagar Yadav](mailto:sagar.yadav@unacademy.com) \
[Khyati](mailto:sai.sankam@unacademy.com)