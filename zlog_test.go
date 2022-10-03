package zlog

// benchmark memory for simple pointer including struct

//func BenchmarkGetLogger(b *testing.B) {
//	var buf bytes.Buffer
//	tt := zerolog.New(&buf).With().Logger()
//	logger := &ZLogger{&tt}
//	for i := 0; i < b.N; i++ {
//		p := logger.Info()
//		p.Int("test", 1).Send()
//		result := buf.String()
//		buf.Reset()
//		if result != `{"_zllevel":"info","test":1}
//` {
//			b.Error("Unexpected result: " + result)
//		}
//	}
//}
