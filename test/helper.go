package test

import (
	"crypto"

	"github.com/crpt/go-crpt"
	"github.com/crpt/go-crpt/factory"
	"github.com/daotl/go-marsha"
	"github.com/daotl/go-marsha/cborgen"

	"github.com/daotl/go-doubl/util"
)

var (
	Mrsh marsha.Marsha = cborgen.New()
	Crpt               = factory.MustNew(crpt.Ed25519, crypto.SHA3_256)
	Util               = util.New(Mrsh, Crpt)
)
