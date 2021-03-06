/**
 * Copyright (C) 2015 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package main

import (
  "bytes"
  "encoding/json"
  "fmt"
  // Dependencies of rbac
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/apimachinery/pkg/api/resource"
  apimachineryversion "k8s.io/apimachinery/pkg/version"
  kapi "k8s.io/api/core/v1"

  apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

  "log"
  "reflect"
  "strings"
  "time"

  "os"

  "github.com/fabric8io/kubernetes-client/kubernetes-model/pkg/schemagen"
)

type Schema struct {
  Info                                     apimachineryversion.Info
  APIGroup                                 metav1.APIGroup
  APIGroupList                             metav1.APIGroupList
  BaseKubernetesList                       metav1.List
  ObjectMeta                               metav1.ObjectMeta
  TypeMeta                                 metav1.TypeMeta
  Status                                   metav1.Status
  Patch                                    metav1.Patch
  ListOptions                              metav1.ListOptions
  DeleteOptions                            metav1.DeleteOptions
  CreateOptions                            metav1.CreateOptions
  UpdateOptions                            metav1.UpdateOptions
  GetOptions                               metav1.GetOptions
  PatchOptions                             metav1.PatchOptions
  Time                                     metav1.Time
  RootPaths                                metav1.RootPaths
  Quantity                                 resource.Quantity
  ObjectReference                          kapi.ObjectReference

  CustomResourceDefinition                 apiextensions.CustomResourceDefinition
  CustomResourceDefinitionList             apiextensions.CustomResourceDefinitionList
  CustomResourceDefinitionSpec             apiextensions.CustomResourceDefinitionSpec
  CustomResourceDefinitionNames            apiextensions.CustomResourceDefinitionNames
  CustomResourceDefinitionCondition        apiextensions.CustomResourceDefinitionCondition
  CustomResourceDefinitionStatus           apiextensions.CustomResourceDefinitionStatus
  // Added JSONSchemaPropsorStringArray here because of
  // https://github.com/joelittlejohn/jsonschema2pojo/issues/866
  JSONSchemaPropsorStringArray             apiextensions.JSONSchemaPropsOrStringArray
}

func main() {
  customTypeNames := map[string]string{
    "K8sSubjectAccessReview": "SubjectAccessReview",
    "K8sLocalSubjectAccessReview":  "LocalSubjectAccessReview",
    "JSONSchemaPropsorStringArray": "JSONSchemaPropsOrStringArray",
  }
  packages := []schemagen.PackageDescriptor{
    {"k8s.io/apimachinery/pkg/util/intstr", "", "io.fabric8.kubernetes.api.model", "kubernetes_apimachinery_pkg_util_intstr_"},
    {"k8s.io/apimachinery/pkg/runtime", "", "io.fabric8.kubernetes.api.model.runtime", "kubernetes_apimachinery_pkg_runtime_"},
    {"k8s.io/apimachinery/pkg/version", "", "io.fabric8.kubernetes.api.model.version", "kubernetes_apimachinery_pkg_version_"},
    {"k8s.io/apimachinery/pkg/apis/meta/v1", "", "io.fabric8.kubernetes.api.model", "kubernetes_apimachinery_"},
    {"k8s.io/api/core/v1", "", "io.fabric8.kubernetes.api.model", "kubernetes_core_"},
    {"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1", "", "io.fabric8.kubernetes.api.model.apiextensions", "kubernetes_apiextensions_"},
  }

  typeMap := map[reflect.Type]reflect.Type{
    reflect.TypeOf(time.Time{}): reflect.TypeOf(""),
    reflect.TypeOf(struct{}{}):  reflect.TypeOf(""),
  }
  schema, err := schemagen.GenerateSchema(reflect.TypeOf(Schema{}), packages, typeMap, customTypeNames, "apiextensions")
  if err != nil {
    fmt.Fprintf(os.Stderr, "An error occurred: %v", err)
    return
  }

  args := os.Args[1:]
  if len(args) < 1 || args[0] != "validation" {
    schema.Resources = nil
  }

  b, err := json.Marshal(&schema)
  if err != nil {
    log.Fatal(err)
  }
  result := string(b)
  result = strings.Replace(result, "\"additionalProperty\":", "\"additionalProperties\":", -1)

  /**
   * Hack to fix https://github.com/fabric8io/kubernetes-client/issues/1565 and https://github.com/fabric8io/kubernetes-client/issues/2144
   *
   * The source golang code uses a type JSON which has a custom serializer and deserializer to ensure that it gets properly
   * translated to and from a string. This gets compiled into a JSON.java class which does not have any such serialization support.
   * This JSON type sounds a lot like JsonNode, which encapsulates any json value. Use that instead.
   */
  result = strings.Replace(result, "{\"$ref\":\"#/definitions/kubernetes_apiextensions_JSON\",\"javaType\":\"io.fabric8.kubernetes.api.model.apiextensions.JSON\"}",
  "{\"javaType\":\"com.fasterxml.jackson.databind.JsonNode\"}", -1)

  var out bytes.Buffer
  err = json.Indent(&out, []byte(result), "", "  ")
  if err != nil {
    log.Fatal(err)
  }

  fmt.Println(out.String())
}
