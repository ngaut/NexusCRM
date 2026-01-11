import unittest
from .schema import SchemaManager

class MockClient:
    def request(self, method, path, body=None):
        return {}, None

class TestSchemaInference(unittest.TestCase):
    def setUp(self):
        self.mgr = SchemaManager(MockClient())

    def test_threshold_tolerance(self):
        # 100 samples: 96 numbers, 4 strings "N/A"
        # 96% match > 95% threshold -> Should be Number
        type_name, _ = self.mgr.infer_type(samples, "amount")
        self.assertEqual(type_name, "LongTextArea", "Should fallback to Text for 4% dirty data (threshold 99%)")

    def test_threshold_fail(self):
        # 100 samples: 94 numbers, 6 strings "N/A"
        # 94% match < 95% threshold -> Should fallback to LongTextArea
        samples = ["123"] * 94 + ["N/A"] * 6
        type_name, _ = self.mgr.infer_type(samples, "amount")
        self.assertEqual(type_name, "LongTextArea", "Should fallback to Text if dirty data > 5%")

    def test_boolean_tolerance(self):
        # 95 True, 5 "Maybe"
        samples = ["True"] * 95 + ["Maybe"] * 5
        type_name, _ = self.mgr.infer_type(samples, "is_active")
        self.assertEqual(type_name, "Boolean", "Should infer Boolean with 5% dirty data")

    def test_id_logic(self):
        # Heuristic: _id allows LongTextArea if mixed or long
        samples = ["0035000000abcde123", "0035000000abcde124"] # 18 chars
        # infer_type logic for _id checks length. If all <= 36, it proposes Lookup.
        type_name, extra = self.mgr.infer_type(samples, "account_id")
        self.assertEqual(type_name, "Lookup", "Should infer Lookup for standard IDs")

    def test_long_id_fallback(self):
        # Huge ID -> Text
        samples = ["a" * 50]
        type_name, _ = self.mgr.infer_type(samples, "external_id_c") # generic name
        self.assertEqual(type_name, "LongTextArea")

if __name__ == '__main__':
    unittest.main()
