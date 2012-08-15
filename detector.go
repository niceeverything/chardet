package chardet

import (
	"errors"
	"sort"
)

type Result struct {
	Charset    string
	Language   string
	Confidence int
}

type Detector struct {
	recognizers []recognizer
}

// List of charset recognizers
var recognizers = []recognizer{
	newRecognizer_utf8(),
	newRecognizer_utf16be(),
	newRecognizer_utf16le(),
	newRecognizer_utf32be(),
	newRecognizer_utf32le(),
	newRecognizer_8859_1_en(),
	newRecognizer_8859_1_da(),
	newRecognizer_8859_1_de(),
	newRecognizer_8859_1_es(),
	newRecognizer_8859_1_fr(),
	newRecognizer_8859_1_it(),
	newRecognizer_8859_1_nl(),
	newRecognizer_8859_1_no(),
	newRecognizer_8859_1_pt(),
	newRecognizer_8859_1_sv(),
	newRecognizer_8859_2_cs(),
	newRecognizer_8859_2_hu(),
	newRecognizer_8859_2_pl(),
	newRecognizer_8859_2_ro(),
	newRecognizer_8859_5_ru(),
	newRecognizer_8859_6_ar(),
	newRecognizer_8859_7_el(),
	newRecognizer_8859_8_I_he(),
	newRecognizer_8859_8_he(),
	newRecognizer_windows_1251(),
	newRecognizer_windows_1256(),
	newRecognizer_KOI8_R(),
	newRecognizer_8859_9_tr(),

	newRecognizer_sjis(),
	newRecognizer_gb_18030(),
	newRecognizer_euc_jp(),
	newRecognizer_euc_kr(),
	newRecognizer_big5(),

	newRecognizer_2022JP(),
	newRecognizer_2022KR(),
	newRecognizer_2022CN(),

	newRecognizer_IBM424_he_rtl(),
	newRecognizer_IBM424_he_ltr(),
	newRecognizer_IBM420_ar_rtl(),
	newRecognizer_IBM420_ar_ltr(),
}

func NewDetector() *Detector {
	return &Detector{recognizers}
}

var (
	NotDetectedError = errors.New("Charset not detected.")
)

func (d *Detector) DetectBest(b []byte, stripTag bool, declaredCharset string) (r *Result, err error) {
	var all []Result
	if all, err = d.DetectAll(b, stripTag, declaredCharset); err == nil {
		r = &all[0]
	}
	return
}

func (d *Detector) DetectAll(b []byte, stripTag bool, declaredCharset string) ([]Result, error) {
	input := newRecognizerInput(b, stripTag, declaredCharset)
	outputChan := make(chan recognizerOutput)
	for _, r := range d.recognizers {
		go matchHelper(r, input, outputChan)
	}
	outputs := make([]recognizerOutput, 0, len(d.recognizers))
	for i := 0; i < len(d.recognizers); i++ {
		o := <-outputChan
		if o.Confidence > 0 {
			outputs = append(outputs, o)
		}
	}
	if len(outputs) == 0 {
		return nil, NotDetectedError
	}

	sort.Sort(recognizerOutputs(outputs))
	dedupOutputs := make([]Result, 0, len(outputs))
	foundCharsets := make(map[string]struct{}, len(outputs))
	for _, o := range outputs {
		if _, found := foundCharsets[o.Charset]; !found {
			dedupOutputs = append(dedupOutputs, Result(o))
			foundCharsets[o.Charset] = struct{}{}
		}
	}
	if len(dedupOutputs) == 0 {
		return nil, NotDetectedError
	}
	return dedupOutputs, nil
}

func matchHelper(r recognizer, input *recognizerInput, outputChan chan<- recognizerOutput) {
	outputChan <- r.Match(input)
}

type recognizerOutputs []recognizerOutput

func (r recognizerOutputs) Len() int           { return len(r) }
func (r recognizerOutputs) Less(i, j int) bool { return r[i].Confidence > r[j].Confidence }
func (r recognizerOutputs) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
