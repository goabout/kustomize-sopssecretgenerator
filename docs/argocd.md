
Use a kustomize [patchStrategicMerge](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/glossary.md#patchstrategicmerge) to apply the following patches to patch the ArgoCD `install.yaml`.

### sopsSecretsGenerator.yaml

    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: argocd-repo-server
    spec:
      template:
        spec:
          volumes:
            - name: custom-tools
              emptyDir: {}
          initContainers:
            - name: install-goaboutsops
              image: alpine:3.11.3
              command: ["/bin/sh", "-c"]
              args:
                - echo "Installing goabout kustomize sops...";
                  set -e;
                  set -x;
                  wget -O /custom-tools/SopsSecretGenerator https://github.com/goabout/kustomize-sopssecretgenerator/releases/download/v${VERSION}/SopsSecretGenerator_${VERSION}_${PLATFORM}_${ARCH};
                  chmod -v +x /custom-tools/SopsSecretGenerator;
                  set +x;
                  echo "Done.";
              volumeMounts:
                - mountPath: /custom-tools
                  name: custom-tools
              env:
                - name: VERSION
                  value: 1.3.1
                - name: PLATFORM
                  value: linux
                - name: ARCH
                  value: amd64
          containers:
            - name: argocd-repo-server
              volumeMounts:
                - mountPath: /.config/kustomize/plugin/goabout.com/v1beta1/sopssecretgenerator/SopsSecretGenerator
                  name: custom-tools
                  subPath: SopsSecretGenerator
              env:
                - name: XDG_CONFIG_HOME
                  value: /.config

### enablePlugins.yaml

    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: argocd-cm
    data:
      kustomize.buildOptions: --enable_alpha_plugins

## Key Managment Server Credendials Examples

*Don't forget to include approprately named secrets with your GCP Service Account, Azure Service Principal (not documented here), or AWS Id and Key*

### gcpServiceAccount.yaml

    apiVersion: apps/v1
    kind: Deployment
    metadata:
    name: argocd-repo-server
    spec:
    template:
      spec:
      volumes:
        - name: gcpServiceAccount
          secret:
            secretName: gcpServiceAccount.json
      containers:
        - name: argocd-repo-server
        volumeMounts:
          - mountPath: /.secrets/gcp/ServiceAccount.json
            name: gcpServiceAccount
            subPath: gcpServiceAccount.json
        env:
          - name: GOOGLE_APPLICATION_CREDENTIALS
            value: /.secrets/gcp/ServiceAccount.json


### awsServiceAccount.yaml

    apiVersion: apps/v1
    kind: Deployment
    metadata:
    name: argocd-repo-server
    spec:
    template:
      spec:
      containers:
        - name: argocd-repo-server
          env:
          - name: AWS_ACCESS_KEY_ID
            valueFrom:
              secretKeyRef:
                name: argocd-aws-credentials
                key: accesskey
          - name: AWS_SECRET_ACCESS_KEY
            valueFrom:
              secretKeyRef:
                name: argocd-aws-credentials
                key: secretkey
