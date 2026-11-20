import json
import os
from pathlib import Path
import unittest

from bridge import InputError, bounded_template, decode_values, read_sheet


def minimal():
    return {"pageDimensions":[100,100],"bubbleDimensions":[10,10],"fieldBlocks":{"answers":{"fieldLabels":["q1..2"],"fieldType":"QTYPE_MCQ4","origin":[10,10],"bubblesGap":15,"labelsGap":20}}}


class TemplateBoundaryTests(unittest.TestCase):
    def test_resource_paths_are_replaced_with_uploaded_resources(self):
        template=minimal()
        template["preProcessors"]=[{"name":"CropOnMarkers","options":{"relativePath":"../../private-file"}},{"name":"FeatureBasedAlignment","options":{"reference":"/private-file"}}]
        safe=bounded_template(template,True,True)
        self.assertEqual(safe["preProcessors"][0]["options"]["relativePath"],"marker.png")
        self.assertEqual(safe["preProcessors"][1]["options"]["reference"],"reference.png")
        self.assertEqual(template["preProcessors"][0]["options"]["relativePath"],"../../private-file")
        with self.assertRaises(InputError): bounded_template(template,False,False)
        template=minimal();template['customLabels']={'Roll':['roll1..999999999999']}
        with self.assertRaises(InputError): bounded_template(template,False,False)
        template=minimal();template['outputColumns']=['q1..999999999999']
        with self.assertRaises(InputError): bounded_template(template,False,False)

    def test_oversized_ranges_and_kernels_are_rejected_before_upstream_import(self):
        for label in ["q1..999999999999", "q8..2"]:
            with self.subTest(label=label):
                template=minimal();template["fieldBlocks"]["answers"]["fieldLabels"]=[label]
                with self.assertRaises(InputError): bounded_template(template,False,False)
        template=minimal();template["preProcessors"]=[{"name":"CropPage","options":{"morphKernel":[1000000,1000000]}}]
        with self.assertRaises(InputError): bounded_template(template,False,False)
        template["preProcessors"]=[{"name":"ArbitraryPlugin","options":{}}]
        with self.assertRaises(InputError): bounded_template(template,False,False)

    def test_multiple_values_and_ambiguous_concatenation(self):
        self.assertEqual(decode_values("AC",list("ABCD")),["A","C"])
        self.assertEqual(decode_values("",list("ABCD")),[])
        self.assertIsNone(decode_values("AB",["A","AB","B"]))
        self.assertIsNone(decode_values("Z",list("ABCD")))

    @unittest.skipUnless(os.environ.get("RUN_OMRCHECKER_TESTS")=="1","Opt-in: installed upstream engine")
    def test_installed_engine_reads_all_original_sample_fields(self):
        root=Path(os.environ.get("OMRCHECKER_ROOT",str(Path.home()/"OMRChecker")))
        sample=Path(__file__).resolve().parents[1]/"apps/web/public/samples/omrchecker"
        source=(sample/"template.json").read_bytes()
        result=read_sheet(root,sample/"filled-sheet.png",{"template":json.loads(source)})
        expected=json.loads((sample/"expected-answers.json").read_text())["responses"]
        self.assertEqual(result["responses"],expected)
        self.assertEqual(result["engine"],"OMRChecker")
        self.assertEqual(len(result["fields"]),26)
        self.assertEqual([f["label"] for f in result["fields"] if f["uncertain"]],["q7","q12"])
        self.assertTrue(result["multiple_marked"])
        self.assertEqual((sample/"template.json").read_bytes(),source)


if __name__=="__main__": unittest.main()
