SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- Database: `tcb38`
--

-- --------------------------------------------------------

--
-- Table structure for table `smartalarm_emails`
--

CREATE TABLE `smartalarm_emails` (
  `id` int(11) NOT NULL COMMENT 'Unique ID for each entry',
  `register_date` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Date that the email address was registered',
  `email` varchar(40) NOT NULL COMMENT 'Email address of the user',
  `type` enum('REGISTERED','CONFIRMED','UNSUBSCRIBED') NOT NULL DEFAULT 'REGISTERED' COMMENT 'Type of the registration',
  `token` varchar(64) NOT NULL COMMENT 'Token for email validation',
  `uses` int(11) NOT NULL DEFAULT '0' COMMENT 'Number of times this email address has been used',
  `last_use` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COMMENT='A list of emails for use with smartalarm';

--
-- Indexes for table `smartalarm_emails`
--
ALTER TABLE `smartalarm_emails`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `email_2` (`email`),
  ADD KEY `email` (`email`);

--
-- AUTO_INCREMENT for table `smartalarm_emails`
--
ALTER TABLE `smartalarm_emails`
  MODIFY `id` int(11) NOT NULL AUTO_INCREMENT COMMENT 'Unique ID for each entry', AUTO_INCREMENT=3354;
  
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
