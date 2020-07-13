package builddata

import (
	"gopkg.in/mgo.v2/bson"
)

type BuildData struct {
	ID                      bson.ObjectId `bson:"_id,omitempty"`
	Jobname                 string
	Buildnumber             int
	Result                  string
	Host                    string
	Duration                int
	Duration_in_queue       float64
	Start                   int
	Waiting                 int
	Workspace               string
	Description             string
	Timediff                int
	Machine                 string
	GerritChangeInfo        bson.M
	GitChangeInfo           bson.M
	Time_build_sh           float64
	Time_bitbake            float64
	Time_rm_BUILD           float64
	Time_rm_BUILD_ARTIFACTS float64
	Time_rm_downloads       float64
	Time_rm_sstate          float64
	Time_rsync_artifacts    float64
	Cause                   bson.M
	Parameters              bson.M
}
