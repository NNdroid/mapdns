package common

var (
	Version   = "v1.0.0"
	GitHash   = ""
	BuildTime = ""
	GoVersion = ""
	Banner    = `

                                                                      
Banner Pic
`
)

func PrintVersion() {
	println(Banner)
	println("Version: ", Version)
	println("GitHash: ", GitHash)
	println("BuildTime: ", BuildTime)
	println("GoVersion: ", GoVersion)
}
