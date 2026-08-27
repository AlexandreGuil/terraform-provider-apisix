resource "apisix_ssl_certificate" "example" {
  certificate = file("example.crt")
  private_key = file("example.key")
  type        = "server"
  labels = {
    "version" = "v1"
  }
}

resource "apisix_ssl_certificate" "mtls" {
  certificate = file("server.crt")
  private_key = file("server.key")
  snis        = ["mtls.example.com"]

  # Require and verify a client certificate signed by the given CA (mutual TLS).
  client = {
    ca                  = file("client-ca.crt")
    depth               = 5
    skip_mtls_uri_regex = ["/public/.*"]
  }
}