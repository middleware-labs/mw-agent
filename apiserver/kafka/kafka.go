package kafka

/*
func consume(server string) {
	// to consume messages

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{server},
		Topic:   "my-topic",
		GroupID: "consumer-group-id",
		//Partition: 0,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			break
		}
		if err != nil {
			break
		}
		log.Printf(strconv.FormatInt(time.Now().Unix(), 10)+" message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}

	if err := reader.Close(); err != nil {
		log.Fatal("failed to close reader:", err)
	}
	fmt.Printf("message read closed")
}

func produce(server string) {
	topic := "my-topic"
	partition := 0

	conn, err := kafka.DialLeader(context.Background(), "tcp", server, topic, partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	for {
		_, err = conn.WriteMessages(
			kafka.Message{Value: []byte("HI " + strconv.FormatInt(time.Now().Unix(), 10))},
			//kafka.Message{Value: []byte("two!")},
			//kafka.Message{Value: []byte("three!")},
		)
		log.Print("message write " + strconv.FormatInt(time.Now().Unix(), 10))
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}

	log.Print("kafka added..")

}*/
