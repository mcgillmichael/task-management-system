-- Insert data into the 'task' table
INSERT INTO task (id, title, description, completed, created_at, updated_at)
VALUES
    (1, 'Upgrade Network Infrastructure', 'Upgrade the network infrastructure to improve performance and reliability.', false, '2023-06-10 09:00:00', '2023-06-10 09:00:00'),
    (2, 'Implement 5G Technology', 'Plan and deploy 5G technology to provide faster and more advanced mobile services.', false, '2023-06-10 10:30:00', '2023-06-10 10:30:00'),
    (3, 'Improve Customer Support System', 'Enhance the customer support system to provide better assistance and resolution to customer issues.', true, '2023-06-09 14:15:00', '2023-06-10 11:45:00');

-- Insert data into the 'task_item' table
INSERT INTO task_item (task_id, item)
VALUES
    (1, 'Evaluate current network equipment'),
    (1, 'Procure and install new routers and switches'),
    (1, 'Configure network devices'),
    (1, 'Perform network testing and optimization'),
    (2, 'Develop a 5G deployment strategy'),
    (2, 'Upgrade base stations and antennas'),
    (2, 'Test 5G network coverage and performance'),
    (2, 'Launch 5G services for customers'),
    (3, 'Review existing support processes'),
    (3, 'Implement a ticketing system'),
    (3, 'Train support staff on new system'),
    (3, 'Monitor and measure customer satisfaction');
