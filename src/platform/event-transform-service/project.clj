(defproject event-transform-service "1.0.0"
  :description "Event Transform Service — transforms and routes events"
  :dependencies [[org.clojure/clojure "1.12.0"]
                 [ring/ring-core "1.12.0"]
                 [ring/ring-jetty-adapter "1.12.0"]
                 [ring/ring-json "0.5.1"]
                 [cheshire "5.13.0"]]
  :main event-transform-service.core
  :aot [event-transform-service.core]
  :uberjar-name "event-transform-service.jar"
  :target-path "target/%s"
  :profiles {:uberjar {:aot :all}})
