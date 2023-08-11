package rpminfo

import (
	"errors"
	"fmt"

	"github.com/antandros/go-pkgparser"
	"github.com/antandros/go-pkgparser/model"
	_ "github.com/glebarez/go-sqlite"
	rpmdb "github.com/knqyf263/go-rpmdb/pkg"
)

func Packages() ([]*rpmdb.PackageInfo, error) {
	db, err := rpmdb.Open("/var/lib/rpm/rpmdb.sqlite")
	if err != nil {
		return nil, err
	}
	pkgList, err := db.ListPackages()
	if err != nil {
		return nil, err
	}
	return pkgList, nil

}
func Parse(p *pkgparser.Parser) error {
	packages, err := Packages()
	if err != nil {
		return err
	}

	for _, rpmPackage := range packages {
		packageItem := p.CreateModel()

		mapItems := map[string]interface{}{
			"Package":         rpmPackage.Name,
			"Version":         rpmPackage.Version,
			"Description":     rpmPackage.Summary,
			"Installed-Size":  rpmPackage.Size,
			"Architecture":    rpmPackage.Arch,
			"License":         rpmPackage.License,
			"Release":         rpmPackage.Release,
			"Modularitylabel": rpmPackage.Modularitylabel,
			"Status":          "installed",
			"Vendor":          rpmPackage.Vendor,
		}
		for key, valn := range mapItems {
			packageItem, err = p.SetValue(key, fmt.Sprintf("%v", valn), packageItem)
			if err != nil {
				fmt.Println("Error", err.Error(), key)
			}
		}
		p.Packages = append(p.Packages, packageItem)

	}
	return nil
}

func GetPackages() ([]model.Package, error) {
	var packages []model.Package
	p := new(pkgparser.Parser)
	p.Model = model.Package{}
	err := p.StructParse()
	if err != nil {
		return nil, err
	}
	err = Parse(p)
	if err != nil {
		return nil, err
	}
	for _, i := range p.Packages {
		item, ok := i.(*model.Package)
		if !ok {
			return nil, errors.New("struct conversion failed")

		}
		packages = append(packages, *item)
	}
	return packages, nil
}
