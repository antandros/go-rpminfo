package rpminfo

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/antandros/go-pkgparser"
	"github.com/antandros/go-pkgparser/model"
	_ "github.com/mattn/go-sqlite3"
)

type PrimaryDatabase struct {
	db     *sql.DB
	dbpath string
}

func OpenPrimaryDB(path string) (*PrimaryDatabase, error) {
	// open database file
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	// TODO: Validate primary_db on open, maybe with the db_info table

	return &PrimaryDatabase{
		db:     db,
		dbpath: path,
	}, nil
}

type PackageEntrySize struct {
	Package   int64 `xml:"type,attr"`
	Installed int64 `xml:"installed,attr"`
	Archive   int64 `xml:"archive,attr"`
}
type PackageEntryVersion struct {
	Epoch   int    `xml:"epoch,attr"`
	Version string `xml:"ver,attr"`
	Release string `xml:"rel,attr"`
}
type PackageEntryLocation struct {
	Href string `xml:"href,attr"`
}
type PackageEntry struct {
	db *PrimaryDatabase

	Key         int
	Arch        string               `xml:"arch"`
	Size        PackageEntrySize     `xml:"size"`
	Location    PackageEntryLocation `xml:"location"`
	PackageName string               `xml:"name"`
	Versions    PackageEntryVersion  `xml:"version"`
	Summary     string               `xml:"summary"`
	Url         string               `xml:"url"`
	License     string               `xml:"rpm_license"`
	Packager    string               `xml:"packager"`
	Vendor      string               `xml:"vendor"`
}
type PackageEntries []PackageEntry

const sqlSelectPackages = `SELECT
 pkgKey
 , name
 , arch
 , epoch
 , version
 , release
 , size_package
 , size_installed
 , size_archive
 , location_href
 , rpm_license
 , rpm_vendor
FROM packages;`

func Packages() (PackageEntries, error) {
	c, err := OpenPrimaryDB("/var/lib/rpm/rpmdb.sqlite")
	if err != nil {
		return nil, err
	}
	rows, err := c.db.Query(sqlSelectPackages)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// parse each row as a package
	packages := make(PackageEntries, 0)
	for rows.Next() {
		p := PackageEntry{
			db: c,
		}

		// scan the values into the slice
		if err = rows.Scan(&p.Key, &p.PackageName, &p.Arch, &p.Versions.Epoch, &p.Versions.Version, &p.Versions.Release, &p.Size.Package, &p.Size.Installed, &p.Size.Archive, &p.Location.Href, &p.License, &p.Vendor); err != nil {
			return nil, fmt.Errorf("error scanning packages: %v", err)
		}

		packages = append(packages, p)
	}

	return packages, nil
}
func Parse(p *pkgparser.Parser) error {
	packages, err := Packages()
	if err != nil {
		return err
	}

	for _, rpmPackage := range packages {
		packageItem := p.CreateModel()

		mapItems := map[string]interface{}{
			"Package":        rpmPackage.PackageName,
			"Version":        rpmPackage.Versions.Version,
			"Description":    rpmPackage.Summary,
			"Installed-Size": rpmPackage.Size.Installed,
			"Architecture":   rpmPackage.Arch,
			"License":        rpmPackage.License,
			"Homepage":       rpmPackage.Url,
			"Revision":       rpmPackage.Versions.Release,
			"Status":         "installed",
			"Vendor":         rpmPackage.Vendor,
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
