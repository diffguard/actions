package lockfile

type Package struct {
	Ecosystem string
	Name      string
	Version   string
	Integrity string
}

func (p Package) VersionKey() string {
	return p.Ecosystem + ":" + p.Name + ":" + p.Version
}

func (p Package) HashKey() string {
	if p.Integrity == "" {
		return ""
	}
	return p.Ecosystem + ":" + p.Name + ":" + p.Integrity
}
