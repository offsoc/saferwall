package goyara

import (
	"path"
	"testing"
)

const (
	yaraRulesPath = "/opt/yararules/"
)

func TestYara(t *testing.T) {
	t.Run("TestYaraLoadRules", func(t *testing.T) {
		rules := []Rule{
			{
				Namespace: "capabilities",
				Filename:  path.Join(yaraRulesPath, "capabilities/capabilities.yar"),
			}, {
				Namespace: "crypto",
				Filename:  path.Join(yaraRulesPath, "crypto/crypto_signatures.yar"),
			},
		}
		_, err := Load(rules)
		if err != nil {
			t.Fatal("failed to load yara rules with error :", err)
		}

	})
	t.Run("TestYaraScanFile", func(t *testing.T) {
		rules := []Rule{
			{
				Namespace: "capabilities",
				Filename:  path.Join(yaraRulesPath, "capabilities/capabilities.yar"),
			}, {
				Namespace: "crypto",
				Filename:  path.Join(yaraRulesPath, "crypto/crypto_signatures.yar"),
			},
		}
		r, err := Load(rules)
		if err != nil {
			t.Fatal("failed to load yara rules with error :", err)
		}
		_, err = ScanFile(r, "../../testdata/putty.exe")
		if err != nil {
			t.Fatal("failed to scan file with error :", err)
		}
	})
}
