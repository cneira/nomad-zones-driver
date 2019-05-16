# First start consul
#consul agent -dev -bind 0.0.0.0 -client 0.0.0.0  &
# Now nomad agent
nomad agent -dev -config=config.hcl -data-dir=$GOPATH/src/github.com/hashicorp/nomad-zone-driver -plugin-dir=$GOPATH/src/github.com/hashicorp/nomad-zones-driver/plugin -bind=0.0.0.0
