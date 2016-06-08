package oebuildjobs

import "encoding/xml"

type Account struct {
	Name     string `xml:"name"`
	Email    string `xml:"email"`
	Username string `xml:"username"`
}

type Provider struct {
	Name    string `xml:"name"`
	Host    string `xml:"host"`
	Port    int    `xml:"port"`
	Scheme  string `xml:"scheme"`
	Url     string `xml:"url"`
	Version string `xml:"version"`
}

type GerritChange struct {
	Project       string  `xml:"project"`
	Branch        string  `xml:"branch"`
	Id            string  `xml:"id"`
	Number        int     `xml:"number"`
	Subject       string  `xml:"subject"`
	CommitMessage string  `xml:"commitMessage"`
	Owner         Account `xml:"owner"`
	Url           string  `xml:"url"`
}

type PatchSet struct {
	Number    int     `xml:"number"`
	Revision  string  `xml:"revision"`
	Ref       string  `xml:"ref"`
	Draft     bool    `xml:"draft"`
	Uploader  Account `xml:"uploader"`
	Author    Account `xml:"Author"`
	Parents   string  `xml:"parents>string"`
	CreatedOn string  `xml:"createdOn"`
}

type Files []string

type Approval struct {
	Type  string `xml:"type"`
	Value string `xml:"value"`
}

type TEvent struct {
	Provider     Provider     `xml:"provider"`
	Account      Account      `xml:"account"`
	GerritChange GerritChange `xml:"change"`
	PatchSet     PatchSet     `xml:"patchSet"`
	Files        Files        `xml:"files>string"`
	Comment      string       `xml:"comment"`
	Approvals    []Approval   `xml:"approvals>com.sonyericsson.hudson.plugins.gerrit.gerritevents.dto.attr.Approval"`
}

type TriggeredItem struct {
	BuildNumber int    `xml:"buildNumber"`
	ProjectId   string `xml:"projectId"`
}

type RetriggerContext struct {
	ThisBuild TriggeredItem   `xml:"thisBuild"`
	Others    []TriggeredItem `xml:"others>triggeredItemEntity"`
}

type RetriggerEvent struct {
	Context RetriggerContext `xml:"context"`
}

type EachParameter struct {
	Name  string `xml:"name"`
	Value string `xml:"value"`
}

type commonBuildAttr struct {
	StartTime   int    `xml:"startTime"`
	Duration    int    `xml:"duration"`
	Result      string `xml:"result"`
	Host        string `xml:"builtOn"`
	Description string `xml:"description"`
}

type VerifyBuild struct {
	XMLName xml.Name `xml:"build"`
	commonBuildAttr
	BuildEvent    TEvent          `xml:"actions>com.sonyericsson.hudson.plugins.gerrit.trigger.hudsontrigger.BadgeAction>tEvent"`
	RetriggerInfo RetriggerEvent  `xml:"actions>com.sonyericsson.hudson.plugins.gerrit.trigger.hudsontrigger.actions.RetriggerAction"`
	Parameters    []EachParameter `xml:"actions>hudson.model.ParametersAction>parameters>hudson.model.StringParameterValue"`
}

type CauseAction struct {
	XMLName xml.Name
	Content []byte `xml:",innerxml"`
}

type Causes struct {
	Causes []CauseAction `xml:",any"`
}

type OfficialBuild struct {
	XMLName xml.Name `xml:"build"`
	commonBuildAttr
	Causes    Causes `xml:"actions>hudson.model.CauseAction>causes"`
	BuildData struct {
		XMLName xml.Name
		Content string `xml:",innerxml"`
	} `xml:"actions>hudson.plugins.git.util.BuildData"`
}
