(ns event-transform-service.core
  (:require [ring.adapter.jetty :as jetty]
            [ring.middleware.json :as json]
            [ring.middleware.params :as params]
            [cheshire.core :as ch])
  (:gen-class))

(defn healthz-handler [_req]
  {:status 200
   :headers {"Content-Type" "application/json"}
   :body (ch/generate-string {:status "ok"})})

(defn transform-handler [req]
  (let [body (get req :body-params {})]
    {:status 200
     :headers {"Content-Type" "application/json"}
     :body (ch/generate-string {:transformed true :input body})}))

(defn router [req]
  (case (:uri req)
    "/healthz" (healthz-handler req)
    "/transform" (transform-handler req)
    {:status 404 :headers {} :body "not found"}))

(def app
  (-> router
      (json/wrap-json-body {:keywords? true})
      (json/wrap-json-response)
      (params/wrap-params)))

(defn -main [& _args]
  (let [port (Integer/parseInt (or (System/getenv "PORT") "50354"))]
    (println (str "event-transform-service starting on port " port))
    (jetty/run-jetty app {:port port :join? true})))
