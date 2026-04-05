package staticfiles

import (
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// URLPrefix é o caminho HTTP onde os arquivos de UPLOADS_PATH ficam disponíveis.
const URLPrefix = "/uploads"

// Mount registra o serviço de arquivos estáticos: URLPrefix -> diretório em diskRoot.
func Mount(router *gin.Engine, diskRoot string) {
	if diskRoot == "" {
		diskRoot = "./uploads"
	}
	abs, err := filepath.Abs(diskRoot)
	if err != nil {
		log.Printf("staticfiles: UPLOADS_PATH inválido %q: %v", diskRoot, err)
		return
	}
	if err := os.MkdirAll(abs, 0755); err != nil {
		log.Printf("staticfiles: não foi possível criar/acessar %s: %v", abs, err)
	}
	router.Static(URLPrefix, abs)
	log.Printf("staticfiles: servindo %s em http://<host>%s/", abs, URLPrefix)
}
