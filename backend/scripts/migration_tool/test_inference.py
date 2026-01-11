
import unittest
from backend.scripts.migration_tool.schema import SchemaManager

class MockClient:
    pass

class TestSchemaInference(unittest.TestCase):
    def setUp(self):
        self.mgr = SchemaManager(MockClient())

    def test_boolean_inference(self):
        # 100 samples: 95 True, 5 "Maybe" -> Should be Boolean (95%)
        data = ["true"] * 95 + ["maybe"] * 5
        type_name, _ = self.mgr.infer_type(data, "is_active")
        self.assertEqual(type_name, "Boolean")

        # 100 samples: 90 True, 10 "Maybe" -> Should be LongTextArea (90% < 95%)
        data = ["true"] * 90 + ["maybe"] * 10
        type_name, _ = self.mgr.infer_type(data, "is_active")
        self.assertEqual(type_name, "LongTextArea")

    def test_number_inference(self):
        # 100 samples: 95 ints, 5 strings -> Number
        data = ["123"] * 95 + ["N/A"] * 5
        type_name, _ = self.mgr.infer_type(data, "amount")
        self.assertEqual(type_name, "Number")

        # 100 samples: 95 currency, 5 strings -> Number
        data = ["$1,000.50"] * 95 + ["TBD"] * 5
        type_name, _ = self.mgr.infer_type(data, "revenue")
        self.assertEqual(type_name, "Number")

    def test_date_inference(self):
        # 100 samples: 95 dates, 5 typos -> DateTime
        data = ["2023-01-01"] * 95 + ["Total"] * 5
        type_name, _ = self.mgr.infer_type(data, "created_date")
        self.assertEqual(type_name, "DateTime")

    def test_id_heuristic(self):
        # If it ends in _id, it's a lookup, even if data looks numeric
        data = ["1001", "1002"]
        type_name, extra = self.mgr.infer_type(data, "owner_id")
        self.assertEqual(type_name, "Lookup")
        self.assertEqual(extra["referenceTo"], ["owner"])

        # If it has _id inside but not end, and data is numeric -> Number
        # Wait, my logic says "if match rate > 95% number", return Number.
        # "zoom_info_company_id_c" -> Number.
        data = ["12345"] * 10
        type_name, _ = self.mgr.infer_type(data, "zoom_info_id_c")
        self.assertEqual(type_name, "Number")

        # If data is NOT numeric, it falls back to text
        data = ["ABC"] * 10
        type_name, _ = self.mgr.infer_type(data, "zoom_info_id_c")
        self.assertEqual(type_name, "LongTextArea")

if __name__ == '__main__':
    unittest.main()
