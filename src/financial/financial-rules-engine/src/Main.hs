module Main where

import Data.Aeson (encode, object, (.=))
import Network.HTTP.Types (status200, status404)
import Network.Wai (Application, pathInfo, responseLBS)
import Network.Wai.Handler.Warp (run)
import System.Environment (lookupEnv)
import Text.Read (readMaybe)

main :: IO ()
main = do
  portStr <- lookupEnv "PORT"
  let port = maybe 50360 (\s -> maybe 50360 id (readMaybe s)) portStr
  putStrLn $ "financial-rules-engine starting on port " ++ show port
  run port app

app :: Application
app req respond =
  case pathInfo req of
    ["healthz"] ->
      respond $
        responseLBS
          status200
          [("Content-Type", "application/json")]
          (encode $ object ["status" .= ("ok" :: String)])
    _ ->
      respond $
        responseLBS
          status404
          [("Content-Type", "application/json")]
          (encode $ object ["error" .= ("not found" :: String)])
