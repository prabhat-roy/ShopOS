require 'sinatra'
require 'json'

set :port, (ENV['PORT'] || 8217).to_i
set :bind, '0.0.0.0'

get '/healthz' do
  content_type :json
  { status: 'ok' }.to_json
end

get '/content/:slug' do
  content_type :json
  { slug: params[:slug], content: nil, found: false }.to_json
end

post '/content/sync' do
  content_type :json
  { status: 'sync_queued' }.to_json
end
